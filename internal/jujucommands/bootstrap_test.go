// Copyright 2025 Canonical.

package jujucommands_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	qt "github.com/frankban/quicktest"
	jujucloud "github.com/juju/juju/cloud"
	"go.uber.org/mock/gomock"

	"github.com/canonical/jimm/v3/internal/jujucommands"
	"github.com/canonical/jimm/v3/internal/jujucommands/mocks"
)

const testBootstrapCACertPEM = `-----BEGIN CERTIFICATE-----
MIID/jCCAmagAwIBAgIVANxsMrzsXrdpjjUoxWQm1RCkmWcqMA0GCSqGSIb3DQEB
CwUAMCYxDTALBgNVBAoTBEp1anUxFTATBgNVBAMTDGp1anUgdGVzdGluZzAeFw0y
MDA0MDgwNTI3NTBaFw0zMDA0MDgwNTMyNTBaMCYxDTALBgNVBAoTBEp1anUxFTAT
BgNVBAMTDGp1anUgdGVzdGluZzCCAaIwDQYJKoZIhvcNAQEBBQADggGPADCCAYoC
ggGBAOW4k2bmXXU3tJ8H5AsGkp8ENLJXzU4SCOCB+X0jPQRVpFtywBVD96z+l+qW
ndGLIg5zMQTtZm71CaOw+8Sl03XU0f28Xrjf+FZCAPID1c7NBttUShbu84euFoCS
C8yobj6JzLz7QswvkshYQ7JEZ88UXtVHqg6MGYFdu+cX/dE1jC7aHg9bus/P6bFH
PVFcHVVxNbLy98Id1iB7i0s97H17nu9O7ZKMrAQAX6dfAELAFQVicdN3WpfwNXEj
M2KIrqttpM8s6/57mi9UJFYGdAEDNkJr/dI506VdGLpiqTFhQK6ztfDfY08QbWk6
iJn8vzWvNW8WthmBtEDpv+DL+a5SJSLSAIZn9sbuBBpiX+csZb66fYhKFFIUrIa5
lrjw8yiHJ4kgsEZJSYaAn7guqmOv8clvy1E2JjsOfGycest6+1/mNdMRFgrMxdzD
0M2yZ96zrdfF/QXpi7Hk7jFLzimuujNUpKFv7k+XObQFxeXnoFrYVkj3YT8hhYF0
mGRkAwIDAQABoyMwITAOBgNVHQ8BAf8EBAMCAqQwDwYDVR0TAQH/BAUwAwEB/zAN
BgkqhkiG9w0BAQsFAAOCAYEAd7GrziPRmjoK3HDF10S+5NgoKYvkOuk2jDap2Qaq
ZFjDvrDA2tr6U0FGY+Hz+QfvtgT+YpJB5IvABvSXdq37llwKGsiSOZSrpHyTsOB0
VcZAF3GMk1nHYMr0o1xRV2gm/ax+vUEStj0k2gTs/p57uhKcCUXR0p3PWXKcRj9a
nVf5bdVkt6ghGs7/uEvj303raUFSf5dJ4C9RTgBK2E9/wlBYNyj5vcsshNpz8kt6
RuARM6Boq2EwKkpRlbvImDM8ZJJLwtpijYrx3egfOSEA7Wfwgwn+B359XcosOee5
n5BC62EjaP85cM9HCtp2DwKjNSosxja723qZPY6Z0Y7IVn3JVjgC2kWP6GViwb+v
l9uwx9ASHPz9ilh6gpjgIifOKZYCaBSS9g8VxHpO5Izxj4vi4AX5cebDg3SzDVik
hZuWHpuOT120okoutwuUSU9448cXLGZfoCZjjdMKXmOj8EEec1diDP4mhegYGezD
LQRNNlaY2ajLt0paowf/Xxb8
-----END CERTIFICATE-----
`

func (s *jujucommandsSuite) TestBootstrapCmdParams_Validate(c *qt.C) {
	p := jujucommands.BootstrapCmdParams{}
	c.Assert(p.Validate(), qt.ErrorMatches, ".*cloud \\[and region\\] name cannot be empty.*")

	p.CloudNameAndRegion = "testregion/testcloud"
	c.Assert(p.Validate(), qt.ErrorMatches, ".*controller name name cannot be empty.*")

	p.ControllerName = "my-controller"
	c.Assert(p.Validate(), qt.ErrorMatches, "missing login token refresh URL, this value should be automatically set by JIMM")

	p.LoginTokenRefreshURL = "myurl.com"
	c.Assert(p.Validate(), qt.IsNil)

	p.AgentVersion = "bad version"
	c.Assert(p.Validate(), qt.ErrorMatches, "invalid version \"bad version\"")

	p.AgentVersion = "1.1.1"
	c.Assert(p.Validate(), qt.IsNil)
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_BuildBootstrapCmdArgs(c *qt.C) {
	tests := []struct {
		name   string
		params jujucommands.BootstrapCmdParams
		expect []string
	}{
		{
			name: "all fields set",
			params: jujucommands.BootstrapCmdParams{
				CloudNameAndRegion:   "testregion/testcloud",
				ControllerName:       "my-controller",
				AgentVersion:         "1.1.1",
				LoginTokenRefreshURL: "myurl.com",
				UserConfig: map[string]string{
					"bootstrap-timeout": "1000",
				},
			},
			expect: []string{
				"bootstrap",
				"--config",
				"login-token-refresh-url=myurl.com",
				"--agent-version=1.1.1",
				"--config",
				"bootstrap-timeout=1000",
				"testregion/testcloud",
				"my-controller",
				fmt.Sprintf("--bootstrap-constraints=arch=%s", runtime.GOARCH),
			},
		},
		{
			name: "no agent version",
			params: jujucommands.BootstrapCmdParams{
				CloudNameAndRegion:   "testregion/testcloud",
				ControllerName:       "my-controller",
				AgentVersion:         "",
				LoginTokenRefreshURL: "myurl.com",
				UserConfig: map[string]string{
					"bootstrap-timeout": "1000",
				},
			},
			expect: []string{
				"bootstrap",
				"--config",
				"login-token-refresh-url=myurl.com",
				"--config",
				"bootstrap-timeout=1000",
				"testregion/testcloud",
				"my-controller",
				fmt.Sprintf("--bootstrap-constraints=arch=%s", runtime.GOARCH),
			},
		},
		{
			name: "no agent version and no bootstrap timeout",
			params: jujucommands.BootstrapCmdParams{
				CloudNameAndRegion:   "testregion/testcloud",
				ControllerName:       "my-controller",
				AgentVersion:         "",
				LoginTokenRefreshURL: "myurl.com",
			},
			expect: []string{
				"bootstrap",
				"--config",
				"login-token-refresh-url=myurl.com",
				"testregion/testcloud",
				"my-controller",
				fmt.Sprintf("--bootstrap-constraints=arch=%s", runtime.GOARCH),
			},
		},
	}

	for _, tt := range tests {
		c.Run(tt.name, func(c *qt.C) {
			args := tt.params.BuildBootstrapCmdArgs()
			c.Assert(args, qt.DeepEquals, tt.expect)
		})
	}
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_RunBootstrapCmd_PersonalCloudWritten(c *qt.C) {
	testCtx := c.Context()

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()

	mockRunner := mocks.NewMockRunner(ctrl)

	dir := c.TempDir()

	mockRunner.EXPECT().JujuDataDir().Return(dir).AnyTimes()
	mockRunner.EXPECT().RunJujuCmd(testCtx, gomock.Any()).DoAndReturn(func(ctx context.Context, args []string) (<-chan jujucommands.OutputLine, error) {
		outputCh := make(chan jujucommands.OutputLine, 1)
		close(outputCh)
		return outputCh, nil
	}).AnyTimes()

	cloudCred := jujucloud.NewNamedCredential("bootstrap-credential", jujucloud.CertificateAuthType, map[string]string{
		"server-cert": "server-cert",
		"client-cert": "client-cert",
		"client-key":  "client-key",
	}, false)

	p := jujucommands.BootstrapCmdParams{
		CloudNameAndRegion:   "testcloud/testregion",
		ControllerName:       "my-controller",
		AgentVersion:         "1.1.1",
		LoginTokenRefreshURL: "myurl.com",

		PersonalCloud: jujucloud.Cloud{
			Type: "lxd",
			AuthTypes: jujucloud.AuthTypes{
				jujucloud.CertificateAuthType,
			},
			// Some fake addr.
			Endpoint: "https://127.0.0.1:8443",
			Regions: []jujucloud.Region{
				{
					Name: "default",
					// Some fake addr.
					Endpoint: "https://127.0.0.1:8443",
				},
			},
		},
		CloudCred: cloudCred,
	}

	cmd := jujucommands.NewBootstrapCmd(mockRunner)

	_, store, cleanup, err := cmd.Run(testCtx, p)
	c.Cleanup(func() {
		cleanup()
	})

	c.Assert(err, qt.IsNil)

	// Now we check the cred is added
	personalCloudCred, err := store.CredentialForCloud("testcloud")
	c.Assert(err, qt.IsNil)

	// Check just attributes, if the populated in memory credential is set for the "testcloud",
	// then we're sure the credential to be used is the one we provided for our provided cloud.
	// (Given it also exists in the temp directory)
	c.Assert(
		personalCloudCred.AuthCredentials["testcloud"].Attributes(),
		qt.DeepEquals,
		cloudCred.Attributes(),
	)

	// This was set to a temp dir until cleanup runs. We can use it to check the file exists.
	_, err = os.Stat(filepath.Join(dir, "clouds.yaml"))
	c.Assert(err, qt.IsNil)
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_RunBootstrapCmd_PublicCloudWritten(c *qt.C) {
	testCtx := c.Context()

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	mockRunner := mocks.NewMockRunner(ctrl)
	dir := c.TempDir()
	mockRunner.EXPECT().JujuDataDir().Return(dir).AnyTimes()

	callCounter := 0
	mockRunner.EXPECT().RunJujuCmd(testCtx, gomock.Any()).DoAndReturn(func(ctx context.Context, args []string) (<-chan jujucommands.OutputLine, error) {
		callCounter++
		if callCounter == 1 {
			// This is only called once within bootstrap, so we can be pretty sure the public-clouds.yaml was written.
			c.Assert(args, qt.DeepEquals, []string{"update-public-clouds", "--client"})
		}

		outputCh := make(chan jujucommands.OutputLine, 1)
		close(outputCh)
		return outputCh, nil
	}).AnyTimes()

	cloudCred := jujucloud.NewNamedCredential("bootstrap-credential", jujucloud.AccessKeyAuthType, map[string]string{
		"aws-secret": "my-secret",
	}, false)
	p := jujucommands.BootstrapCmdParams{
		CloudNameAndRegion:   "aws/us-east-1",
		ControllerName:       "my-controller",
		AgentVersion:         "1.1.1",
		LoginTokenRefreshURL: "myurl.com",

		CloudCred: cloudCred,
	}

	cmd := jujucommands.NewBootstrapCmd(mockRunner)

	_, store, cleanup, err := cmd.Run(
		testCtx,
		p,
	)
	c.Cleanup(func() {
		cleanup()
	})

	c.Assert(err, qt.IsNil)

	// Now we check the cred is added for the cloud (excluding region)
	personalCloudCred, err := store.CredentialForCloud("aws")
	c.Assert(err, qt.IsNil)
	c.Assert(
		personalCloudCred.AuthCredentials["aws"].Attributes(),
		qt.DeepEquals,
		cloudCred.Attributes(),
	)

	// Make sure no personal cloud was written.
	_, err = os.Stat(filepath.Join(dir, "clouds.yaml"))
	c.Assert(err, qt.ErrorMatches, ".*no such file or directory.*")
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_RunBootstrapCmd_WritesBootstrapCloudInitConfig(c *qt.C) {
	testCtx := c.Context()

	ctrl := gomock.NewController(c)
	defer ctrl.Finish()
	mockRunner := mocks.NewMockRunner(ctrl)
	dir := c.TempDir()
	mockRunner.EXPECT().JujuDataDir().Return(dir).AnyTimes()

	callCounter := 0
	mockRunner.EXPECT().RunJujuCmd(testCtx, gomock.Any()).DoAndReturn(func(ctx context.Context, args []string) (<-chan jujucommands.OutputLine, error) {
		callCounter++

		// First call is update-public-clouds, second call is bootstrap.
		if callCounter == 2 {
			var cfgPath string
			for i := 0; i < len(args)-1; i++ {
				if args[i] == "--config" && strings.HasSuffix(args[i+1], "bootstrap-cloudinit.yaml") {
					cfgPath = args[i+1]
					break
				}
			}
			c.Assert(cfgPath, qt.Not(qt.Equals), "")
			b, err := os.ReadFile(cfgPath)
			c.Assert(err, qt.IsNil)
			c.Assert(string(b), qt.Contains, "cloudinit-userdata: |")
			c.Assert(string(b), qt.Contains, "ca-certs:")
			c.Assert(string(b), qt.Contains, "-----BEGIN CERTIFICATE-----")
		}

		outputCh := make(chan jujucommands.OutputLine, 1)
		close(outputCh)
		return outputCh, nil
	}).AnyTimes()

	cloudCred := jujucloud.NewNamedCredential("bootstrap-credential", jujucloud.CertificateAuthType, map[string]string{
		"server-cert": "server-cert",
		"client-cert": "client-cert",
		"client-key":  "client-key",
	}, false)

	p := jujucommands.BootstrapCmdParams{
		CloudNameAndRegion:        "testcloud/testregion",
		ControllerName:            "my-controller",
		AgentVersion:              "1.1.1",
		LoginTokenRefreshURL:      "myurl.com",
		BootstrapTrustedCACertPEM: testBootstrapCACertPEM,
		PersonalCloud: jujucloud.Cloud{
			Type: "lxd",
			AuthTypes: jujucloud.AuthTypes{
				jujucloud.CertificateAuthType,
			},
			Endpoint: "https://127.0.0.1:8443",
			Regions: []jujucloud.Region{{
				Name:     "default",
				Endpoint: "https://127.0.0.1:8443",
			}},
		},
		CloudCred: cloudCred,
	}

	cmd := jujucommands.NewBootstrapCmd(mockRunner)
	_, _, cleanup, err := cmd.Run(testCtx, p)
	c.Cleanup(func() {
		cleanup()
	})

	c.Assert(err, qt.IsNil)
	_, err = os.Stat(filepath.Join(dir, "bootstrap-cloudinit.yaml"))
	c.Assert(err, qt.IsNil)
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_Validate_RejectsCloudInitUserDataOverrideWhenCACertSet(c *qt.C) {
	p := jujucommands.BootstrapCmdParams{
		CloudNameAndRegion:        "testcloud/testregion",
		ControllerName:            "my-controller",
		LoginTokenRefreshURL:      "myurl.com",
		BootstrapTrustedCACertPEM: testBootstrapCACertPEM,
		UserConfig: map[string]string{
			"cloudinit-userdata": "some-user-cloudinit",
		},
	}

	err := p.Validate()
	c.Assert(err, qt.IsNotNil)
	c.Assert(err.Error(), qt.Contains, "cloudinit-userdata")
}

func (s *jujucommandsSuite) TestBootstrapCmdParams_Validate_RejectsInvalidCACertPEM(c *qt.C) {
	p := jujucommands.BootstrapCmdParams{
		CloudNameAndRegion:        "testcloud/testregion",
		ControllerName:            "my-controller",
		LoginTokenRefreshURL:      "myurl.com",
		BootstrapTrustedCACertPEM: "not-a-pem",
		UserConfig:                map[string]string{},
	}

	err := p.Validate()
	c.Assert(err, qt.IsNotNil)
	c.Assert(err.Error(), qt.Contains, "invalid bootstrap trusted CA cert PEM")
}
