package integration

import (
	"fmt"
	"github.com/microsoft/cobalt/test-harness/infratests"
	"github.com/microsoft/cobalt/test-harness/terratest-extensions/modules/azure"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

// Verifies that the webhook configured in ACR is configured to use the Webhook URI exported
// by the webapp. If this criteria is not met, the webhook in the ACR won't target the
// webapp effectively.
func verifyCorrectWebhookEndpointForApps(goTest *testing.T, output infratests.TerraformOutput) {
	acrName := output["acr_name"].(string)
	appDevResourceGroup := output["app_dev_resource_group"].(string)
	adminResourceGroup := output["admin_resource_group"].(string)

	for _, appName := range output["webapp_names"].([]interface{}) {
		appNameS := appName.(string)
		acrWebHook := azure.ACRWebHookCallback(
			goTest,
			adminSubscription,
			appDevResourceGroup,
			acrName,
			getWebhookNameForWebApp(output, appNameS))

		cdURI := azure.WebAppCDUri(goTest, adminSubscription, adminResourceGroup, appNameS)
		require.Equal(goTest, cdURI, *acrWebHook.ServiceURI)
	}
}

// Verifies that the webapp is configured to deploy the correct image. Without this validation we cannot
// be sure that the CD webhook that will trigger the deployment is going to properly target the
// correct webapp.
func verifyCorrectDeploymentTargetForApps(goTest *testing.T, output infratests.TerraformOutput) {
	adminResourceGroup := output["admin_resource_group"].(string)
	acrName := output["acr_name"].(string)

	for appIndex, appName := range output["webapp_names"].([]interface{}) {
		appConfig := azure.WebAppSiteConfiguration(goTest, adminSubscription, adminResourceGroup, appName.(string))
		linuxFxVersion := *appConfig.LinuxFxVersion

		var expectedImageName string = ""
		var expectedImageTagPrefix string = ""

		for deploy_target := range unauthn_deploymentTargets.([]map[string]string) {
			if strings.Contains(linuxFxVersion, fmt.Sprintf("%s:%s-%s", deploy_target["image_name"], deploy_target["image_release_tag_prefix"], workspace)) {
				expectedImageName = deploy_target["image_name"]
				expectedImageTagPrefix = deploy_target["image_release_tag_prefix"]
			}
		}

		for deploy_target := range authn_deploymentTargets.([]map[string]string) {
			if strings.Contains(linuxFxVersion, fmt.Sprintf("%s:%s-%s", deploy_target["image_name"], deploy_target["image_release_tag_prefix"], workspace)) {
				expectedImageName = deploy_target["image_name"]
				expectedImageTagPrefix = deploy_target["image_release_tag_prefix"]
			}
		}

		expectedAcr := acrName + ".azurecr.io"
		expectedLinuxFxVersion := fmt.Sprintf(
			"DOCKER|%s/%s:%s-%s",
			expectedAcr,
			expectedImageName,
			expectedImageTagPrefix,
			workspace)

		require.Equal(goTest, expectedLinuxFxVersion, linuxFxVersion)
	}
}
