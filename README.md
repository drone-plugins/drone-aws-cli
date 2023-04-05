# AWS CLI Plugin

The aws-cli plugin is designed to automate the installation, configuration, and execution of the AWS Command Line Interface (CLI) in your CI pipeline. This plugin simplifies the process of setting up and managing AWS credentials and configurations within your pipeline, allowing you to easily interact with AWS services.

Features of the plugin include:

1. Installation of a specified version of the AWS CLI or the latest version if not specified.
2. Automatic configuration of AWS credentials and profiles based on the provided environment variables.
3. Support for AWS session tokens for temporary credentials and enhanced security.
4. Optional configuration of the default and profile-specific AWS regions. 
5. Customization of the AWS CLI installation directory and binary directory. 
6. Optional disabling of the AWS CLI output paging. 

By incorporating this plugin into your pipeline, you can effortlessly interact with AWS services without the need to manually configure and manage the AWS CLI. This makes it more convenient and secure to deploy and manage resources on AWS within your CI/CD workflows.

# Usage

The following settings changes this plugin's behavior.

* aws_access_key_id (required) sets the AWS Access Key ID.
* aws_secret_access_key (required) sets the AWS Secret Access Key.
* aws_region (required) sets the AWS Region.
* aws_session_token (optional) sets the AWS Session Token.
* binary_dir (optional) sets the binary directory for the AWS CLI. Default: /usr/local/bin.
* disable_aws_pager (optional) controls AWS CLI output paging. Default: true.
* install_dir (optional) sets the installation directory for the AWS CLI. Default: /usr/local/aws-cli.
* install_dir (optional) sets the installation directory for the AWS CLI. Default: /usr/local/aws-cli.
* override_installed (optional) controls whether to override the installed AWS CLI version. Default: false.
* profile_name (optional) sets the profile name to be configured. Default: default.
* role_arn (optional) sets the Role ARN for assuming an IAM role with web identity.
* role_session_name (optional) sets the Role Session Name for assuming an IAM role with web identity.
* session_duration (optional) sets the Session Duration for assuming an IAM role with web identity. Default: 3600.
* version (optional) sets the AWS CLI version to be installed. Default: latest.
* configure_default_region (optional) controls whether to configure the default region. Default: true.
* configure_profile_region (optional) controls whether to configure the profile region. Default: true.

Below is an example `.drone.yml` that uses this plugin.

```yaml
kind: pipeline
name: default

steps:
  - name: run plugins/aws-cli plugin
    image: plugins/aws-cli
    pull: if-not-exists
    settings:
      aws_access_key_id: your_aws_access_key_id
      aws_secret_access_key: your_aws_secret_access_key
      aws_region: your_aws_region
      aws_session_token: your_aws_session_token
      binary_dir: /usr/local/bin
      disable_aws_pager: true
      install_dir: /usr/local/aws-cli
      override_installed: false
      profile_name: default
      role_arn: your_role_arn
      role_session_name: your_role_session_name
      session_duration: "3600"
      version: latest
      configure_default_region: true
      configure_profile_region: true
```

# Building

Build the plugin binary:

```text
scripts/build.sh
```

Build the plugin image:

```text
docker build -t plugins/aws-cli -f docker/Dockerfile .
```

# Testing

Execute the plugin from your current working directory:

```text
docker run --rm \
  -e PLUGIN_AWS_ACCESS_KEY_ID=your_aws_access_key_id \
  -e PLUGIN_AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key \
  -e PLUGIN_AWS_REGION=your_aws_region \
  -e PLUGIN_PROFILE_NAME=your_aws_profile_name \
  -e PLUGIN_AWS_SESSION_TOKEN=your_aws_session_token \
  -e DRONE_COMMIT_SHA=8f51ad7884c5eb69c11d260a31da7a745e6b78e2 \
  -e DRONE_COMMIT_BRANCH=master \
  -e DRONE_BUILD_NUMBER=43 \
  -e DRONE_BUILD_STATUS=success \
  -w /drone/src \
  -v $(pwd):/drone/src \
  plugins/aws-cli

```
