// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

package setup

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const awsInfo = `
HouseKeeper can be used to monitor your AWS account
in order to keep resource usage and cost down for you.

Running this setup will give HouseKeeper access to
EC2 and S3 in order to perform monitoring and cleanup.
`

// This allows a role to be assumed by the houskeeper user in the shared QA AWS account
const awsAssumeRoleDoc = `{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Principal": {
				"AWS": "arn:aws:iam::475063612724:user/jenkins-housekeeper"
			},
			"Action": "sts:AssumeRole"
		}
	]
  }`

const (
	roleName   = "brkt-HouseKeeper"
	policyName = "HouseKeeperPolicy"
	policyDesc = "Allow HouseKeeper to access your resources"

	awsPolicyOrRoleExist = "EntityAlreadyExists"

	awsPolicyARNTemplate = "arn:aws:iam::%s:policy/%s"

	awsIDKey     = "AWS_ACCESS_KEY_ID"
	awsSecretKey = "AWS_SECRET_ACCESS_KEY"
)

var (
	monitorEC2 = []string{"ec2:DescribeInstances", "ec2:DescribeInstanceAttribute", "ec2:DescribeSnapshots", "ec2:DescribeVolumeStatus", "ec2:DescribeVolumes", "ec2:DescribeInstanceStatus", "ec2:DescribeTags", "ec2:DescribeVolumeAttribute", "ec2:DescribeImages", "ec2:DescribeSnapshotAttribute"}
	monitorS3  = []string{"s3:GetBucketTagging", "s3:ListBucket", "s3:GetObject", "s3:ListAllMyBuckets", "s3:GetBucketLocation"}

	cleanupEC2 = []string{"ec2:DeregisterImage", "ec2:DeleteSnapshot", "ec2:DeleteTags", "ec2:ModifyImageAttribute", "ec2:DeleteVolume", "ec2:TerminateInstances", "ec2:CreateTags", "ec2:StopInstances"}
	cleanupS3  = []string{"s3:PutBucketTagging", "s3:DeleteObject", "s3:DeleteBucket"}

	errPolicyExist = errors.New("A policy with the same name already exist")
	errRoleExist   = errors.New("A role with the same name already exist")
	errSkipAWS     = errors.New("Don't override AWS settings")
)

func awsSetup() error {
	fmt.Println("Performing AWS setup...")

	_, idExist := os.LookupEnv(awsIDKey)
	_, secretExist := os.LookupEnv(awsSecretKey)
	if !idExist || !secretExist {
		return errors.New("No AWS credentials exist")
	}
	fmt.Println(awsInfo)

	// Get user preferences
	conf := getAWSConf()
	if !conf.cleanup && !conf.monitor {
		fmt.Println("Skipping AWS setup...")
		return nil
	}

	sess := session.Must(session.NewSession())
	iamClient := iam.New(sess, &aws.Config{})

	// First create a policy based on what the user configured
	policy, err := createAWSPolicy(policyName, conf, iamClient)
	if err == errPolicyExist {
		// A policy already exist create a new policy with random suffix
		rand.Seed(time.Now().UnixNano())
		newPolicyName := fmt.Sprintf("%s-%d", policyName, rand.Int63())
		policy, err = createAWSPolicy(newPolicyName, conf, iamClient)
		if err != nil {
			return fmt.Errorf("Failed to create policy: %s", err)
		}
	} else if err != nil {
		return err
	}
	fmt.Printf("Created new policy:\n\t%s\n", *policy.Arn)

	// Now create role
	role, err := createAWSRole(roleName, iamClient)
	if err == errRoleExist {
		// A role already exist, replace it
		err = deleteAWSRole(roleName, iamClient)
		if err != nil {
			return fmt.Errorf("Failed to delete old role: %s", err)
		}
		role, err = createAWSRole(roleName, iamClient)
		if err != nil {
			return fmt.Errorf("Could not create HouseKeeper role: %s", err)
		}
	} else if err != nil {
		return err
	}
	fmt.Printf("Created new role:\n\t%s\n", *role.Arn)

	// Finally connect the policy to the role
	_, err = iamClient.AttachRolePolicy(&iam.AttachRolePolicyInput{
		RoleName:  (*role).RoleName,
		PolicyArn: (*policy).Arn,
	})
	if err != nil {
		return fmt.Errorf("Could not attach HouseKeeper policy to HouseKeeper role: %s", err)
	}
	return nil
}

func getAWSConf() *config {
	conf := new(config)
	if !getYes("Allow HouseKeeper to monitor and cleanup?", true) {
		return conf
	}
	// Don't let the user choose what to allow, per request
	conf.monitorEC2 = true
	conf.monitorS3 = true
	conf.cleanupEC2 = true
	conf.cleanupS3 = true
	conf.monitor = true
	conf.cleanup = true
	/*
		conf.monitor = getYes("Setup monitoring (read) of your resources?", true)
		if conf.monitor {
			conf.monitorEC2 = getYes("\tMonitor EC2:", true)
			conf.monitorS3 = getYes("\tMonitor S3", true)
		}
		if !conf.monitor {
			return conf
		}
		conf.cleanup = getYes("Setup cleanup (read & write) of your resources?", true)
		if conf.cleanup {
			conf.cleanupEC2 = getYes("\tCleanup EC2:", true)
			conf.cleanupS3 = getYes("\tCleanup S3", true)
		}
	*/
	return conf
}

func deleteAWSRole(name string, iamClient *iam.IAM) error {
	// First detach all attached policies
	out, err := iamClient.ListAttachedRolePolicies(&iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(name),
	})
	if err != nil {
		return err
	}
	for _, pol := range out.AttachedPolicies {
		_, err := iamClient.DetachRolePolicy(&iam.DetachRolePolicyInput{
			RoleName:  aws.String(name),
			PolicyArn: pol.PolicyArn,
		})
		if err != nil {
			return err
		}
	}

	// Now actually delete the role
	_, err = iamClient.DeleteRole(&iam.DeleteRoleInput{
		RoleName: aws.String(name),
	})
	return err
}

func createAWSRole(name string, iamClient *iam.IAM) (*iam.Role, error) {
	// Create a new role that can be assumed by HouseKeeper
	input := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(awsAssumeRoleDoc),
		Description:              aws.String(policyDesc),
		RoleName:                 aws.String(roleName),
	}
	out, err := iamClient.CreateRole(input)
	if err != nil {
		// Role might already exist
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == awsPolicyOrRoleExist {
			return nil, errRoleExist
		}
		// Other error
		return nil, err
	}
	return out.Role, nil
}

func createAWSPolicy(name string, conf *config, iamClient *iam.IAM) (*iam.Policy, error) {
	input := &iam.CreatePolicyInput{
		Description:    aws.String(policyDesc),
		PolicyName:     aws.String(name),
		PolicyDocument: aws.String(conf.PolicyJSON()),
	}
	out, err := iamClient.CreatePolicy(input)
	if err != nil {
		// Could be that the policy already exist
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == awsPolicyOrRoleExist {
			return nil, errPolicyExist
		}
		// Other error
		return nil, err
	}
	// Can't return (out.Policy, err) directly as out might == nil if err != nil
	return out.Policy, nil
}

type config struct {
	monitor, monitorEC2, monitorS3 bool
	cleanup, cleanupEC2, cleanupS3 bool
}

func (c config) String() string {
	template := `

### HouseKeeper configuration ###

Allow monitoring of resources: %t
	EC2:	%t
	S3: 	%t

Allow cleanup of resource: %t
	EC2:	%t
	S3: 	%t

`
	return fmt.Sprintf(template, c.monitor, c.monitorEC2, c.monitorS3, c.cleanup, c.cleanupEC2, c.cleanupS3)
}

type policyStatement struct {
	Sid      string
	Effect   string
	Action   []string
	Resource string
}

type policyDocument struct {
	Version   string
	Statement []policyStatement
}

func (c config) Policy() policyDocument {
	actionSet := make(map[string]struct{})
	if c.monitorEC2 || c.cleanupEC2 {
		for i := range monitorEC2 {
			actionSet[monitorEC2[i]] = struct{}{}
		}
	}
	if c.monitorS3 || c.cleanupS3 {
		for i := range monitorS3 {
			actionSet[monitorS3[i]] = struct{}{}
		}
	}
	if c.cleanupEC2 {
		for i := range cleanupEC2 {
			actionSet[cleanupEC2[i]] = struct{}{}
		}
	}
	if c.cleanupS3 {
		for i := range cleanupS3 {
			actionSet[cleanupS3[i]] = struct{}{}
		}
	}

	doc := policyDocument{}
	statement := policyStatement{}
	statement.Action = []string{}

	for action := range actionSet {
		statement.Action = append(statement.Action, action)
	}

	statement.Effect = "Allow"
	statement.Resource = "*"
	statement.Sid = "VisualEditor0"

	doc.Version = "2012-10-17"
	doc.Statement = []policyStatement{statement}
	return doc
}

func (c config) PolicyJSON() string {
	doc := c.Policy()
	b, err := json.Marshal(doc)
	if err != nil {
		log.Fatalln("Failed to encode AWS policy")
	}
	return string(b)
}
