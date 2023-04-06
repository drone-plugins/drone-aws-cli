// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// AWS Access Key ID.
	AWSAccessKeyID string `envconfig:"PLUGIN_AWS_ACCESS_KEY_ID"`

	// AWS Secret Access Key.
	AWSSecretAccessKey string `envconfig:"PLUGIN_AWS_SECRET_ACCESS_KEY"`

	// AWS Session Token.
	AWSSessionToken string `envconfig:"PLUGIN_AWS_SESSION_TOKEN"`

	// AWS Region.
	AWSRegion string `envconfig:"PLUGIN_AWS_REGION"`

	// Binary directory for the AWS CLI.
	BinaryDir string `envconfig:"PLUGIN_BINARY_DIR"`

	// Disable AWS CLI output paging.
	DisableAWSPager bool `envconfig:"PLUGIN_DISABLE_AWS_PAGER"`

	// Installation directory for the AWS CLI.
	InstallDir string `envconfig:"PLUGIN_INSTALL_DIR"`

	// Override the installed AWS CLI version.
	OverrideInstalled bool `envconfig:"PLUGIN_OVERRIDE_INSTALLED"`

	// Profile name to be configured.
	ProfileName string `envconfig:"PLUGIN_PROFILE_NAME"`

	// Role ARN.
	RoleARN string `envconfig:"PLUGIN_ROLE_ARN"`

	// Role session name.
	RoleSessionName string `envconfig:"PLUGIN_ROLE_SESSION_NAME"`

	// Session duration.
	SessionDuration string `envconfig:"PLUGIN_SESSION_DURATION"`

	// AWS CLI version.
	Version string `envconfig:"PLUGIN_VERSION"`

	// Configure default region.
	ConfigureDefaultRegion bool `envconfig:"PLUGIN_CONFIGURE_DEFAULT_REGION"`

	// Configure profile region.
	ConfigureProfileRegion bool `envconfig:"PLUGIN_CONFIGURE_PROFILE_REGION"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	// Set default value for args.BinaryDir if not provided
	log.Println("Starting AWS CLI plugin")
	if args.InstallDir == "" {
		if runtime.GOOS == "windows" {
			args.InstallDir = `C:\Program Files\Amazon\AWSCLI`
		} else {
			args.InstallDir = "/usr/local/aws-cli"
		}
	}
	if args.BinaryDir == "" {
		if runtime.GOOS == "windows" {
			args.BinaryDir = `C:\Program Files\Amazon\AWSCLI\bin`
		} else {
			args.BinaryDir = "/usr/local/bin"
		}
	}
	err := installAWSCLI(args)
	log.Printf("AWS CLI binary path: %s\n", args.BinaryDir)

	os.Setenv("PATH", fmt.Sprintf("%s:%s", args.BinaryDir, os.Getenv("PATH")))

	if err != nil {
		return fmt.Errorf("Error: %v\n", err)
	}
	err = configureCredentials(args)
	if err != nil {
		return fmt.Errorf("Error: %v\n", err)
	}
	// Assume role with web identity if the Role ARN and Role Session Name are provided.
	if args.RoleARN != "" && args.RoleSessionName != "" {
		err = assumeRoleWithWebIdentity(args)
		if err != nil {
			return fmt.Errorf("Error: %v\n", err)
		}
	}

	log.Println("AWS CLI plugin completed successfully")
	return nil
}

func configureCredentials(args Args) error {
	if args.AWSAccessKeyID == "" {
		return fmt.Errorf("AWS access key ID not provided")
	}
	if args.AWSSecretAccessKey == "" {
		return fmt.Errorf("AWS secret access key not provided")
	}

	// Configure AWS credentials
	configureCredentialsCmd := exec.Command("aws", "configure", "set", "aws_access_key_id", args.AWSAccessKeyID, "--profile", args.ProfileName)
	err := configureCredentialsCmd.Run()
	if err != nil {
		return err
	}
	configureCredentialsCmd = exec.Command("aws", "configure", "set", "aws_secret_access_key", args.AWSSecretAccessKey, "--profile", args.ProfileName)
	err = configureCredentialsCmd.Run()
	if err != nil {
		return err
	}
	if args.AWSSessionToken != "" {
		configureCredentialsCmd = exec.Command("aws", "configure", "set", "aws_session_token", args.AWSSessionToken, "--profile", args.ProfileName)
		err = configureCredentialsCmd.Run()
		if err != nil {
			return err
		}
	}

	// Configure default region if specified.
	if args.ConfigureDefaultRegion {
		if args.AWSRegion == "" {
			return fmt.Errorf("AWS region not provided")
		}
		configureRegionCmd := exec.Command("aws", "configure", "set", "default.region", args.AWSRegion, "--profile", args.ProfileName)
		err = configureRegionCmd.Run()
		if err != nil {
			return err
		}
	}

	// Configure profile region if specified.
	if args.ConfigureProfileRegion {
		if args.AWSRegion == "" {
			return fmt.Errorf("AWS region not provided")
		}
		configureProfileRegionCmd := exec.Command("aws", "configure", "set", "region", args.AWSRegion, "--profile", args.ProfileName)
		err = configureProfileRegionCmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func installAWSCLI(args Args) error {
	log.Println("Installing AWS CLI")
	if !args.OverrideInstalled {
		_, err := exec.LookPath("aws")
		if err == nil {
			fmt.Println("AWS CLI is already installed. Skipping installation.")
			return nil
		}
	}

	// Installation process
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("sh", "-c", `curl -sSL "https://awscli.amazonaws.com/AWSCLIV2$1.pkg" -o "AWSCLIV2.pkg"
			sudo installer -pkg AWSCLIV2.pkg -target /
			rm AWSCLIV2.pkg`, args.Version)
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			cmd = exec.Command("sh", "-c", `curl -sSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64$1.zip" -o "awscliv2.zip"
				unzip -q -o awscliv2.zip
				sudo ./aws/install -i "${PLUGIN_INSTALL_DIR}" -b "${PLUGIN_BINARY_DIR}"
				rm -r awscliv2.zip ./aws`, args.Version)
		case "arm64":
			cmd = exec.Command("sh", "-c", `curl -sSL "https://awscli.amazonaws.com/awscli-exe-linux-aarch64$1.zip" -o "awscliv2.zip"
				unzip -q -o awscliv2.zip
				sudo ./aws/install -i "${PLUGIN_INSTALL_DIR}" -b "${PLUGIN_BINARY_DIR}"
				rm -r awscliv2.zip ./aws`, args.Version)
		default:
			return fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	log.Println("AWS CLI installed successfully")

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PLUGIN_INSTALL_DIR=%s", args.InstallDir))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PLUGIN_BINARY_DIR=%s", args.BinaryDir))

	err := cmd.Run()
	if err != nil {
		return err
	}

	if args.DisableAWSPager {
		disablePagerCmd := exec.Command("aws", "configure", "set", "cli_pager", "")
		err = disablePagerCmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func assumeRoleWithWebIdentity(args Args) error {
	assumeRoleCmd := exec.Command("aws", "sts", "assume-role-with-web-identity",
		"--role-arn", args.RoleARN,
		"--role-session-name", args.RoleSessionName,
		"--duration-seconds", args.SessionDuration,
		"--profile", args.ProfileName)

	output, err := assumeRoleCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to assume role with web identity: %v\n%s", err, string(output))
	}

	fmt.Printf("Assumed role with web identity:\n%s\n", string(output))

	return nil
}
