package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "9df601cb-78de-44bc-a49a-1f36f0b220ef"

func TestAzureLinuxVMCreation(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "sawh0007",
		},
	}

	defer terraform.Destroy(t, terraformOptions)

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the value of output variables
	vmName := terraform.Output(t, terraformOptions, "vm_name")               // "sawh0007"
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name") // "sawh0007-A05-RG"
	nicName := terraform.Output(t, terraformOptions, "nic_name")             // "sawh0007A05Nic"
	publicIPName := terraform.Output(t, terraformOptions, "public_ip_name")  // "sawh0007A05PublicIP"
	nsgName := terraform.Output(t, terraformOptions, "nsg_name")             // "sawh0007A05SG"
	vnetName := terraform.Output(t, terraformOptions, "vnet_name")           // "sawh0007A05Vnet"

	// Confirm VM exists
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID))

	// Confirm NIC exists and is connected to the VM
	assert.True(t, azure.NetworkInterfaceExists(t, nicName, resourceGroupName, subscriptionID))
	nic := azure.GetNetworkInterface(t, nicName, resourceGroupName, subscriptionID)
	assert.Equal(t, nic.VirtualMachine.ID, "/subscriptions/"+subscriptionID+"/resourceGroups/"+resourceGroupName+"/providers/Microsoft.Compute/virtualMachines/"+vmName)

	// Confirm the VM is running the correct Ubuntu version (e.g., Ubuntu 20.04 LTS)
	vm := azure.GetVirtualMachine(t, vmName, resourceGroupName, subscriptionID)
	assert.Contains(t, vm.StorageProfile.OsDisk.OsType, "Linux")
	assert.Equal(t, vm.StorageProfile.OsDisk.ImageReference.Sku, "20.04-LTS")

	// Confirm Public IP exists
	assert.True(t, azure.PublicIPExists(t, publicIPName, resourceGroupName, subscriptionID))

	// Confirm NSG exists and is associated with the NIC
	nsg := azure.GetNetworkSecurityGroup(t, nsgName, resourceGroupName, subscriptionID)
	assert.Contains(t, nsg.SecurityRules, "AllowSSH") // Ensure that the NSG allows SSH connections.

	// Confirm Virtual Network exists
	vnet := azure.GetVirtualNetwork(t, vnetName, resourceGroupName, subscriptionID)
	assert.True(t, vnet.Name != "")

	// Confirm the VM size
	assert.Equal(t, vm.HardwareProfile.VmSize, "Standard_DS1_v2")
}
