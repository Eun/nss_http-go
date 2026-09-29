package libtest

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

type TestContainer struct {
	container  testcontainers.Container
	ConfigFile string
}

func NewTestContainer(configContents string) (*TestContainer, error) {
	var mounts testcontainers.ContainerMounts
	var configFilePath string
	if configContents != "" {
		configFile, err := os.CreateTemp("", "nss_http_")
		if err != nil {
			return nil, errors.Wrap(err, "unable to create temp file")
		}
		if _, err := configFile.WriteString(configContents); err != nil {
			return nil, errors.Wrap(err, "unable to write nss_http.json")
		}
		if err := configFile.Close(); err != nil {
			return nil, errors.Wrap(err, "unable to close temp file")
		}
		mounts = append(mounts, testcontainers.ContainerMount{
			Source:   testcontainers.GenericBindMountSource{HostPath: configFile.Name()},
			Target:   "/etc/nss_http.json",
			ReadOnly: true,
		})
		configFilePath = configFile.Name()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: func() string {
				if name := os.Getenv("TEST_CONTAINER"); name != "" {
					return name
				}
				return "nss_http_test:latest"
			}(),
			ExposedPorts: []string{"22/tcp"},
			Mounts:       mounts,
			HostConfigModifier: func(config *container.HostConfig) {
				config.AutoRemove = true
				// host.docker.internal is only predefined on Docker Desktop
				// (macOS/Windows). On Linux it has to be mapped explicitly,
				// or every HTTP provider lookup fails and NSS silently falls
				// back to the local files, which makes the whole suite fail.
				//
				// This must be set here rather than via ContainerRequest.
				// ExtraHosts: supplying a HostConfigModifier replaces
				// testcontainers' default modifier, which is what would
				// otherwise copy ExtraHosts into the host config.
				config.ExtraHosts = append(config.ExtraHosts,
					"host.docker.internal:host-gateway")
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return &TestContainer{
		container:  c,
		ConfigFile: configFilePath,
	}, nil
}

func (tc *TestContainer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := tc.container.Terminate(ctx); err != nil {
		return errors.Wrap(err, "unable to terminate container")
	}
	if tc.ConfigFile != "" {
		if err := os.Remove(tc.ConfigFile); err != nil {
			return errors.Wrap(err, "unable to delete nss_http.json")
		}
	}
	return nil
}

func (tc *TestContainer) SSHAddr() (string, error) {
	return tc.container.PortEndpoint(context.Background(), "22/tcp", "")
}

func (tc *TestContainer) getent(database, key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := []string{"getent", database}
	if key != "" {
		cmd = append(cmd, key)
	}

	_, r, err := tc.container.Exec(ctx, cmd, exec.Multiplexed())
	if err != nil {
		return "", errors.Wrap(err, "unable to exec in container")
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to read buffer")
	}
	return strings.TrimSpace(string(buf)), nil
}

func (tc *TestContainer) GetPasswd(name string) (string, error) {
	return tc.getent("passwd", name)
}

func (tc *TestContainer) GetShadow(name string) (string, error) {
	return tc.getent("shadow", name)
}

func (tc *TestContainer) GetGroup(name string) (string, error) {
	return tc.getent("group", name)
}

func (tc *TestContainer) GetGShadow(name string) (string, error) {
	return tc.getent("gshadow", name)
}

func (tc *TestContainer) Members(groupName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"members", "-t", groupName}, exec.Multiplexed())
	if err != nil {
		return "", errors.Wrap(err, "unable to exec in container")
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to read buffer")
	}
	return strings.TrimSpace(string(buf)), nil
}

// ID runs "id <user>" in the container, which resolves the user's primary and
// supplementary groups through NSS.
func (tc *TestContainer) ID(userName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"id", userName}, exec.Multiplexed())
	if err != nil {
		return "", errors.Wrap(err, "unable to exec in container")
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to read buffer")
	}
	return strings.TrimSpace(string(buf)), nil
}

func (tc *TestContainer) Groups(userName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"groups", userName}, exec.Multiplexed())
	if err != nil {
		return "", errors.Wrap(err, "unable to exec in container")
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to read buffer")
	}
	return strings.TrimSpace(string(buf)), nil
}

func (tc *TestContainer) GetLogs() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"cat", "/var/log/nss_http.log"}, exec.Multiplexed())
	if err != nil {
		return "", errors.Wrap(err, "unable to exec in container")
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", errors.Wrap(err, "unable to read buffer")
	}
	return strings.TrimSpace(string(buf)), nil
}
