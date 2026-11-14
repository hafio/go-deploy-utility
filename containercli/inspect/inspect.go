// Package inspect asks docker or podman about containers and decodes the answer
// into the fields solmq-conn-util and solace-util read, smoothing over where the two
// engines differ: docker's leading slash on the name, and the health block that
// older podman spells Healthcheck.
//
// The container's environment is never decoded, because it holds secrets, so no
// report built from a Container can print one.
package inspect

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/hafio/go-deploy-utility/runner"
)

// Containers asks for the named containers: `container inspect <names...>`. The
// container subcommand answers for containers alone, so an image or a volume with
// the same name never does. base is the engine argv the tool built and checked with
// execguard, and it is copied. With no names there is nothing to ask, and ok is
// false.
func Containers(base []string, names ...string) (c runner.Cmd, ok bool) {
	if len(names) == 0 {
		return runner.Cmd{}, false
	}
	argv := make([]string, 0, len(base)+2+len(names))
	argv = append(argv, base...)
	argv = append(argv, "container", "inspect")
	return runner.Cmd{Argv: append(argv, names...)}, true
}

// Container is one container as the engine reported it.
type Container struct {
	// Name is the container's name, without docker's leading slash.
	Name string

	// Created is when the container was created, as the engine wrote it.
	Created string

	// Image is the image the container was configured with (Config.Image).
	Image string

	// Labels are the container's labels (Config.Labels).
	Labels map[string]string

	State State

	// RestartCount is the engine's own count. On podman, systemd restarts the
	// container, so a tool takes the count from systemd.NRestarts instead.
	RestartCount int

	// NanoCPUs and Memory are the CPU and memory limits (HostConfig), 0 when unset.
	NanoCPUs, Memory int64

	Mounts []Mount

	// Networks are the names of the networks the container is attached to, sorted.
	Networks []string

	// Secrets are the names of the secrets podman created the container with; docker
	// reports none, since its compose secrets arrive as mounts.
	Secrets []string
}

// State is the container's run state.
type State struct {
	Status                                 string
	Running, Restarting, Paused, OOMKilled bool
	ExitCode                               int
	StartedAt, FinishedAt, Error           string

	// Health is the healthcheck's verdict, zero when the container has none.
	Health Health
}

// Health is a healthcheck's verdict.
type Health struct {
	Status        string
	FailingStreak int
}

// Mount is one of the container's mounts.
type Mount struct {
	Type, Name, Source, Destination string
	RW                              bool
}

// wire is the shape of one element of the engine's answer. encoding/json matches
// the engine's keys to these field names without regard to case.
type wire struct {
	Name, Created string
	State         struct {
		Status                                 string
		Running, Restarting, Paused, OOMKilled bool
		ExitCode                               int
		StartedAt, FinishedAt, Error           string
		Health, Healthcheck                    *Health
	}
	RestartCount int
	Config       struct {
		Image   string
		Labels  map[string]string
		Secrets []struct {
			Name string
		}
	}
	HostConfig struct {
		NanoCpus, Memory int64
	}
	Mounts          []Mount
	NetworkSettings struct {
		Networks map[string]json.RawMessage
	}
	// Secrets is a fallback for an answer that lists them outside Config. Podman
	// reports them under Config (libpod's InspectContainerConfig).
	Secrets []struct {
		Name string
	}
}

// Decode reads Containers' answer, a JSON array, in the order the engine returned
// it. An empty answer, one that is not JSON and one that is not an array are errors;
// an empty array is no containers.
func Decode(out string) ([]Container, error) {
	out = strings.TrimSpace(out)
	if out == "" {
		return nil, errors.New("inspect: the engine answered nothing; check that the container names " +
			"exist and that the engine is reachable")
	}
	var ws []wire
	if err := json.Unmarshal([]byte(out), &ws); err != nil {
		return nil, fmt.Errorf("inspect: the engine's answer is not a list of containers: %w", err)
	}
	cs := make([]Container, 0, len(ws))
	for _, w := range ws {
		c := Container{
			Name:         strings.TrimPrefix(w.Name, "/"),
			Created:      w.Created,
			Image:        w.Config.Image,
			Labels:       w.Config.Labels,
			RestartCount: w.RestartCount,
			NanoCPUs:     w.HostConfig.NanoCpus,
			Memory:       w.HostConfig.Memory,
			Mounts:       w.Mounts,
			State: State{
				Status:     w.State.Status,
				Running:    w.State.Running,
				Restarting: w.State.Restarting,
				Paused:     w.State.Paused,
				OOMKilled:  w.State.OOMKilled,
				ExitCode:   w.State.ExitCode,
				StartedAt:  w.State.StartedAt,
				FinishedAt: w.State.FinishedAt,
				Error:      w.State.Error,
				Health:     health(w.State.Health, w.State.Healthcheck),
			},
		}
		for n := range w.NetworkSettings.Networks {
			c.Networks = append(c.Networks, n)
		}
		sort.Strings(c.Networks)
		secrets := w.Config.Secrets
		if len(secrets) == 0 {
			secrets = w.Secrets
		}
		for _, s := range secrets {
			c.Secrets = append(c.Secrets, s.Name)
		}
		cs = append(cs, c)
	}
	return cs, nil
}

// health is the first of the two spellings that carries a status. An empty Health
// block beside a populated Healthcheck must not read as "no healthcheck".
func health(spellings ...*Health) Health {
	for _, h := range spellings {
		if h != nil && h.Status != "" {
			return *h
		}
	}
	return Health{}
}

// Printable drops every control character and every format character (category
// Cf) from s. Names, statuses, error text and mount paths come from the engine,
// about a container the tool did not necessarily create. A terminal escape inside
// one would be acted on by the terminal that prints it, and a bidi override or a
// zero-width character makes the line read differently from its bytes, which is
// why execguard refuses the same characters in a command.
func Printable(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}
