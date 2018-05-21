package main

// BEFORE RUNNING:
// ---------------
// 0. Change your nameservers to Google Cloud DNS nameservers.
// 1. If not already done, enable the Google Cloud DNS API
//    and check the quota for your project at
//    https://console.developers.google.com/apis/api/dns
// 2. This sample uses Application Default Credentials for authentication.
//    If not already done, install the gcloud CLI from
//    https://cloud.google.com/sdk/ and run
//    `gcloud beta auth application-default login`.
//    For more information, see
//    https://developers.google.com/identity/protocols/application-default-credentials
// 3. Install and update the Go dependencies by running `go get -u` in the
//    project directory.
//
// Usage: IP=1.2.3.4 FQDN=test.something.com. ZONE=YOUR_ZONE PROJECT=YOUR_PROJECT ./google-dns
// Note: Make sure that FQDN ends with ".".

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1"
)

func main() {
	ctx := context.Background()

	c, err := google.DefaultClient(ctx, dns.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	dnsService, err := dns.New(c)
	if err != nil {
		log.Fatal(err)
	}

	project, ok := os.LookupEnv("PROJECT")
	if !ok {
		fmt.Println("Project not present.")
		return
	}

	zone, ok := os.LookupEnv("ZONE")
	if !ok {
		fmt.Println("Zone not present.")
		return
	}

	fqdn, ok := os.LookupEnv("FQDN")
	if !ok {
		fmt.Println("FQDN not present.")
		return
	}

	ip, ok := os.LookupEnv("IP")
	if !ok {
		fmt.Println("IP not present.")
		return
	}

	rec := &dns.ResourceRecordSet{
		Name:    fqdn,
		Rrdatas: []string{ip},
		Ttl:     int64(300),
		Type:    "A",
	}

	rb := &dns.Change{
		Additions: []*dns.ResourceRecordSet{rec},
	}

	list, err := dnsService.ResourceRecordSets.List(project, zone).Name(fqdn).Type("A").Do()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	if len(list.Rrsets) > 0 {
		// Attempt to delete the existing records when adding our new one.
		rb.Deletions = list.Rrsets
	}

	resp, err := dnsService.Changes.Create(project, zone, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Be smarter than this.
	fmt.Printf("%#v\n", resp)
}
