package testhelpers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/exec"
)

type TestContainer struct {
	container          testcontainers.Container
	NSSwitchConfigFile string
	ConfigFile         string
}

func NewTestContainer(nsswitchConfContents, configContents string) (*TestContainer, error) {
	nsswitchFile, err := os.CreateTemp("", "nss_http_")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %w", err)
	}
	if _, err := nsswitchFile.WriteString(nsswitchConfContents); err != nil {
		return nil, fmt.Errorf("unable to write nsswitch.conf: %w", err)
	}
	if err := nsswitchFile.Close(); err != nil {
		return nil, fmt.Errorf("unable to close temp file: %w", err)
	}
	configFile, err := os.CreateTemp("", "nss_http_")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp file: %w", err)
	}
	if _, err := configFile.WriteString(configContents); err != nil {
		return nil, fmt.Errorf("unable to write nss_http.json: %w", err)
	}
	if err := configFile.Close(); err != nil {
		return nil, fmt.Errorf("unable to close temp file: %w", err)
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
			Mounts: testcontainers.Mounts(
				testcontainers.ContainerMount{
					Source:   testcontainers.GenericBindMountSource{HostPath: nsswitchFile.Name()},
					Target:   "/etc/nsswitch.conf",
					ReadOnly: true,
				},
				testcontainers.ContainerMount{
					Source:   testcontainers.GenericBindMountSource{HostPath: configFile.Name()},
					Target:   "/etc/nss_http.json",
					ReadOnly: true,
				},
			),
			HostConfigModifier: func(config *container.HostConfig) {
				config.AutoRemove = true
			},
		},
		Started: true,
	})
	if err != nil {
		return nil, err
	}
	return &TestContainer{
		container:          c,
		NSSwitchConfigFile: nsswitchFile.Name(),
		ConfigFile:         configFile.Name(),
	}, nil
}

func (tc *TestContainer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := tc.container.Terminate(ctx); err != nil {
		return fmt.Errorf("unable to terminate container: %w", err)
	}
	if err := os.Remove(tc.NSSwitchConfigFile); err != nil {
		return fmt.Errorf("unable to delete nsswitch.conf: %w", err)
	}
	if err := os.Remove(tc.ConfigFile); err != nil {
		return fmt.Errorf("unable to delete nss_http.json: %w", err)
	}
	return nil
}

func (tc *TestContainer) GetUser(name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"getent", "passwd", name}, exec.Multiplexed())
	if err != nil {
		return "", fmt.Errorf("unable to exec in container: %w", err)
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("unable to read buffer: %w", err)
	}
	return strings.TrimSpace(string(buf)), nil
}

func (tc *TestContainer) GetLogs() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, r, err := tc.container.Exec(ctx, []string{"cat", "/var/log/nss_http.log"}, exec.Multiplexed())
	if err != nil {
		return "", fmt.Errorf("unable to exec in container: %w", err)
	}
	buf, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("unable to read buffer: %w", err)
	}
	return strings.TrimSpace(string(buf)), nil
}
