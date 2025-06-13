package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "22f60889-0ad3-47d4-9d41-7d026e7ff990"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "du000086",
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variable
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")

	// Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Test NIC connection
	t.Run("Verify NIC Connection", func(t *testing.T) {
		vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)
		assert.NotEmpty(t, *vm.NetworkProfile.NetworkInterfaces, "VM should have at least one network interface")
		nicID := *(*vm.NetworkProfile.NetworkInterfaces)[0].ID
		nicName := azure.GetNameFromResourceID(nicID)
		assert.True(t, azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID), "Network interface should exist")
	})

	// Test Ubuntu version
	t.Run("Verify Ubuntu Version", func(t *testing.T) {
		vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)

		assert.Equal(t, "Canonical", *vm.StorageProfile.ImageReference.Publisher, "Image publisher should be Canonical")
		assert.Equal(t, "0001-com-ubuntu-server-jammy", *vm.StorageProfile.ImageReference.Offer, "Image offer should be Ubuntu Server 22.04 LTS")
		assert.Equal(t, "22_04-lts-gen2", *vm.StorageProfile.ImageReference.Sku, "Image SKU should be 22.04 LTS Gen2")
	})
}
