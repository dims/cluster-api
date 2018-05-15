package clients

import (
	"fmt"
	"os/exec"
	"testing"

	openstackconfigv1 "sigs.k8s.io/cluster-api/cloud/openstack/openstackproviderconfig/v1alpha1"
)

func TestInstanceList(t *testing.T) {
	// cloud config for openlab
	cfg := &CloudConfig{
		Username:   "clustera",
		Password:   "Huan-0615",
		DomainName: "clustera",
		TenantID:   "CH-819764",
		Region:     "",
		AuthUrl:    "https://openstack.tacc.chameleoncloud.org:5000/v2.0",
	}
	is, err := NewInstanceService(cfg)
	if err != nil {
		t.Fatalf("Create new instance service err: %v", err)
	}

	// test update token
	oldToken := is.provider.Token()
	is.provider.SetToken("fake-token")
	err = is.UpdateToken()
	if err != nil {
		t.Fatalf("Test update token err: %v", err)
	}
	if oldToken == is.provider.Token() {
		t.Fatalf("Test failed, expect new token be different with new token")
	}
	if is.provider.Token() == "fake-token" {
		t.Fatalf("Test failed, token didn't update")
	}

	// test create new instance
	config := &openstackconfigv1.OpenstackProviderConfig{
		Name:   "test-01",
		Image:  "CC-Ubuntu16.04-20160610",
		Flavor: "m1.medium",
		Networks: []openstackconfigv1.NetworkParam{{
			UUID: "ab14ce0d-5e1f-4e32-bf65-00416e3cc19c",
		}},
		AvailabilityZone: "nova",
		FloatingIP:       "129.114.111.153",
	}
	fakecmd := `#cloud-config
disable_root: false
runcmd:
- [ bash, -c, "sudo echo "Hello, world!" > /root/lw" ]`
	instance, err := is.InstanceCreate(config, nil, fakecmd, "root-3h")
	if err != nil {
		t.Fatalf("Create instance err: %v", err)
	}
	fmt.Printf("\nInstance create success, instance detail is:\n%+v\n", instance)

	// test instance list
	list, err := is.getInstanceList(nil)
	if err != nil || len(*list) == 0 {
		t.Fatalf("Get instance list failed.")
	}

	for _, instance := range *list {
		detail, err := is.GetInstance(instance.ID)
		if err != nil {
			t.Fatalf("Get instance detail failed: %v", err)
		}
		fmt.Printf("instance detail is: %+v", detail)
	}

	// test extract IP from instance
	ip := ""
	for _, *instance = range *list {
		if instance.Name == "test-01" {
			ip, _ = getIPFromInstance(*instance)
		}
	}
	if ip == "" {
		t.Fatal("Extract IP err")
	}
	fmt.Printf("Got ip is: %q", ip)

	// // test create ssh key pair
	// b, _ := ioutil.ReadFile("/root/.ssh/id_rsa.pub")
	// publicKey := string(b)
	// userName := fmt.Sprintf("root-%s", util.RandomString(2))
	// err = is.CreateKeyPair(userName, publicKey)
	// if err != nil {
	// 	t.Fatalf("Create keypair err")
	// }

	// test use ssh key run command in instance
	cmd := exec.Command(
		"ssh", "-i", "/root/.ssh/id_rsa",
		"-o", "StrictHostKeyChecking no",
		"-o", "UserKnownHostsFile /dev/null",
		fmt.Sprintf("%s@%s", "root-3h", ip),
		"echo STARTFILE; sudo cat /root/lw")
	output, err := cmd.Output()
	if string(output) != "Hello, word!" || err != nil {
		t.Fatalf("exec ssh command err, res is: %s, err is: %+v", string(output), err)
	}
}
