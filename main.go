package main

import (
	"strings"

	"github.com/pulumi/pulumi-azure/sdk/v3/go/azure/compute"
	"github.com/pulumi/pulumi-azure/sdk/v3/go/azure/core"
	"github.com/pulumi/pulumi-azure/sdk/v3/go/azure/network"
	"github.com/pulumi/pulumi-azure/sdk/v3/go/azure/storage"
	"github.com/pulumi/pulumi/sdk/v2/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vmName := "vm-valheim"
		portSlice := []string{"Tcp:3389", "Tcp:2456", "Udp:2456", "Tcp:2457", "Udp:2457", "Tcp:2458", "Udp:2458"}

		// Create an Azure Resource Group
		resourceGroup, err := core.NewResourceGroup(ctx, "rg-valheim", &core.ResourceGroupArgs{
			Location: pulumi.String("WestEurope"),
		})
		if err != nil {
			return err
		}

		// Create Virtual Network
		mainVirtualNetwork, err := network.NewVirtualNetwork(ctx, "vnet-valheim", &network.VirtualNetworkArgs{
			AddressSpaces: pulumi.StringArray{
				pulumi.String("10.0.0.0/16"),
			},
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
		})

		// Create Subnet
		internal, err := network.NewSubnet(ctx, "snet-valheim-10-0-2-0_24", &network.SubnetArgs{
			ResourceGroupName:  resourceGroup.Name,
			VirtualNetworkName: mainVirtualNetwork.Name,
			AddressPrefixes: pulumi.StringArray{
				pulumi.String("10.0.2.0/24"),
			},
		})
		if err != nil {
			return err
		}

		nsg, err := network.NewNetworkSecurityGroup(ctx, "nsg-valheim", &network.NetworkSecurityGroupArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			Tags: pulumi.StringMap{
				"environment": pulumi.String("staging"),
			},
		})
		if err != nil {
			return err
		}

		for _, port := range portSlice {
			portInfo := strings.Split(port, ":")
			portProtocol := portInfo[0]
			portStr := portInfo[1]
			ruleName := portStr + "-" + portProtocol + "-rule"

			_, err = network.NewNetworkSecurityRule(ctx, ruleName, &network.NetworkSecurityRuleArgs{
				Priority:                 pulumi.Int(100),
				Direction:                pulumi.String("Inbound"),
				Access:                   pulumi.String("Allow"),
				Protocol:                 pulumi.String(portProtocol),
				SourcePortRange:          pulumi.String("*"),
				DestinationPortRange:     pulumi.String(portStr),
				SourceAddressPrefix:      pulumi.String("*"),
				DestinationAddressPrefix: pulumi.String("*"),
				ResourceGroupName:        resourceGroup.Name,
				NetworkSecurityGroupName: nsg.Name,
			})
			if err != nil {
				return err
			}
		}

		_, err = network.NewSubnetNetworkSecurityGroupAssociation(ctx, "valheimSubnetNetworkSecurityGroupAssociation", &network.SubnetNetworkSecurityGroupAssociationArgs{
			SubnetId:               internal.ID(),
			NetworkSecurityGroupId: nsg.ID(),
		})
		if err != nil {
			return err
		}

		// Public IP
		valheimPublicIP, err := network.NewPublicIp(ctx, "pip-valheim", &network.PublicIpArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			AllocationMethod:  pulumi.String("Static"),
			Sku:               pulumi.String("Standard"),
		})
		if err != nil {
			return err
		}

		// Create Network Interface
		nicName := "nic-" + vmName
		mainNetworkInterface, err := network.NewNetworkInterface(ctx, nicName, &network.NetworkInterfaceArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			IpConfigurations: network.NetworkInterfaceIpConfigurationArray{
				&network.NetworkInterfaceIpConfigurationArgs{
					Name:                       pulumi.String("testconfiguration1"),
					SubnetId:                   internal.ID(),
					PrivateIpAddressAllocation: pulumi.String("Dynamic"),
					PublicIpAddressId:          valheimPublicIP.ID(),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create Virtual Machine
		_, err = compute.NewVirtualMachine(ctx, vmName, &compute.VirtualMachineArgs{
			Location:          resourceGroup.Location,
			ResourceGroupName: resourceGroup.Name,
			NetworkInterfaceIds: pulumi.StringArray{
				mainNetworkInterface.ID(),
			},
			VmSize: pulumi.String("Standard_DS2_v2"),
			StorageImageReference: &compute.VirtualMachineStorageImageReferenceArgs{
				Publisher: pulumi.String("MicrosoftWindowsServer"),
				Offer:     pulumi.String("WindowsServer"),
				Sku:       pulumi.String("2019-Datacenter"),
				Version:   pulumi.String("latest"),
			},
			StorageOsDisk: &compute.VirtualMachineStorageOsDiskArgs{
				Name:            pulumi.String("myosdisk1"),
				Caching:         pulumi.String("ReadWrite"),
				CreateOption:    pulumi.String("FromImage"),
				ManagedDiskType: pulumi.String("Standard_LRS"),
			},
			OsProfile: &compute.VirtualMachineOsProfileArgs{
				ComputerName:  pulumi.String(vmName),
				AdminUsername: pulumi.String("testadmin"),
				AdminPassword: pulumi.String("Password1234!"),
			},
			OsProfileWindowsConfig: &compute.VirtualMachineOsProfileWindowsConfigArgs{
				Timezone:                pulumi.String("W. Europe Standard Time"),
				EnableAutomaticUpgrades: pulumi.Bool(true),
			},
			Tags: pulumi.StringMap{
				"environment": pulumi.String("staging"),
			},
		})
		if err != nil {
			return err
		}

		// Create an Azure resource (Storage Account)
		account, err := storage.NewAccount(ctx, "savalheim", &storage.AccountArgs{
			ResourceGroupName:      resourceGroup.Name,
			AccountTier:            pulumi.String("Standard"),
			AccountReplicationType: pulumi.String("LRS"),
		})
		if err != nil {
			return err
		}

		/*
			// Create a Load Balancer
			valheimLoadBalancer, err := lb.NewLoadBalancer(ctx, "lb-valheim", &lb.LoadBalancerArgs{
				Location:          resourceGroup.Location,
				ResourceGroupName: resourceGroup.Name,
				Sku:               pulumi.String("Standard"),
				FrontendIpConfigurations: lb.LoadBalancerFrontendIpConfigurationArray{
					&lb.LoadBalancerFrontendIpConfigurationArgs{
						Name:              pulumi.String("PublicIPAddress"),
						PublicIpAddressId: valheimPublicIP.ID(),
					},
				},
			})
			if err != nil {
				return err
			}

			valheimBackendPool, err := lb.NewBackendAddressPool(ctx, "bep-valheim", &lb.BackendAddressPoolArgs{
				LoadbalancerId:    valheimLoadBalancer.ID(),
				ResourceGroupName: resourceGroup.Name,
			})
			if err != nil {
				return err
			}

			_, err = network.NewNetworkInterfaceBackendAddressPoolAssociation(ctx, "bepa-valheim", &network.NetworkInterfaceBackendAddressPoolAssociationArgs{
				NetworkInterfaceId:   mainNetworkInterface.ID(),
				IpConfigurationName:  pulumi.String("testconfiguration1"),
				BackendAddressPoolId: valheimBackendPool.ID(),
			})
			if err != nil {
				return err
			}

			for _, port := range portSlice {
				portInfo := strings.Split(port, ":")

				portStr := portInfo[1]
				portInt, err := strconv.Atoi(portStr)
				portProtocol := portInfo[0]
				probeName := portStr + "-" + portProtocol + "-probe"
				ruleName := portStr + "-" + portProtocol + "-rule"
				probe, err := lb.NewProbe(ctx, probeName, &lb.ProbeArgs{
					ResourceGroupName: resourceGroup.Name,
					LoadbalancerId:    valheimLoadBalancer.ID(),
					Port:              pulumi.Int(portInt),
				})
				if err != nil {
					return err
				}

				_, err = lb.NewRule(ctx, ruleName, &lb.RuleArgs{
					ResourceGroupName:           resourceGroup.Name,
					LoadbalancerId:              valheimLoadBalancer.ID(),
					Protocol:                    pulumi.String(portProtocol),
					FrontendPort:                pulumi.Int(portInt),
					BackendPort:                 pulumi.Int(portInt),
					FrontendIpConfigurationName: pulumi.String("PublicIPAddress"),
					BackendAddressPoolId:        valheimBackendPool.ID(),
					ProbeId:                     probe.ID(),
				})
				if err != nil {
					return err
				}
			}
		*/

		// Export the connection string for the storage account
		ctx.Export("connectionString", account.PrimaryConnectionString)
		ctx.Export("VirtualNetworkName", mainVirtualNetwork.Name)
		ctx.Export("SubnetName", internal.Name)
		ctx.Export("NicName", mainNetworkInterface.ID())
		ctx.Export("PublicIp", valheimPublicIP.IpAddress)
		//ctx.Export("BackendPoolName", valheimBackendPool.Name)
		return nil
	})
}
