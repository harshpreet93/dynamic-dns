package main

import (
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"io/ioutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/aws"
	"time"
)

func initContext() {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.ReadInConfig()
}

func getMyIp() (string, error) {
	resp, err := http.DefaultClient.Get("https://api.ipify.org?format=text")

	if err != nil {
		fmt.Println("get unsuccessful")
		return "", err
	}

	defer resp.Body.Close()

	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}
	return string(response[:]), nil
}

func switchARecordIPAddr(hostedZoneID string, aRecord string, ipAddr string) error {
	sess := session.New()

	svc := route53.New(sess)

	rrs := &route53.ResourceRecordSet{ // Required
		Name: aws.String(aRecord), // Required
		Type: aws.String("A"),  // Required
		ResourceRecords: []*route53.ResourceRecord{{
			Value: &ipAddr,
		}},
		TTL: aws.Int64(300),
	}

	route53Change := route53.Change{ // Required
		Action: aws.String("UPSERT"), // Required
		ResourceRecordSet: rrs,
	}

	allChanges := []*route53.Change{
		&route53Change,
	}

	changeBatch := &route53.ChangeBatch{ // Required
		Changes: allChanges,
		Comment: aws.String("ResourceDescription"),
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: changeBatch,
		HostedZoneId: aws.String(hostedZoneID), // Required
	}
	
	resp, err := svc.ChangeResourceRecordSets(params)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	fmt.Println(resp)

	return nil
}

func main()  {
	initContext()
	for {
		ip, err := getMyIp()
		if err != nil {
			break
		}
		switchARecordIPAddr(viper.Get("hosted-zone-id").(string), viper.Get("a-record").(string), ip)
		fmt.Println("Waiting for 5 minutes")
		time.Sleep(5 * time.Minute)
	}
}
