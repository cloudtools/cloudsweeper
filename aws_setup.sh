#!/bin/bash
set -e
POLICY_NAME="HouseKeeperPolicy"
ROLE_NAME="brkt-HouseKeeper"

HOUSEKEEPER_POLICY='{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceAttribute",
                "ec2:DescribeSnapshots",
                "ec2:DescribeVolumeStatus",
                "ec2:DescribeVolumes",
                "ec2:DescribeInstanceStatus",
                "ec2:DescribeTags",
                "ec2:DescribeVolumeAttribute",
                "ec2:DescribeImages",
                "ec2:DescribeSnapshotAttribute",
                "ec2:DeregisterImage",
                "ec2:DeleteSnapshot",
                "ec2:DeleteTags",
                "ec2:ModifyImageAttribute",
                "ec2:DeleteVolume",
                "ec2:TerminateInstances",
                "ec2:CreateTags",
                "ec2:StopInstances",
                "s3:GetBucketTagging",
                "s3:ListBucket",
                "s3:GetObject",
                "s3:ListAllMyBuckets",
                "s3:GetBucketLocation",
                "s3:PutBucketTagging",
                "s3:DeleteObject",
                "s3:DeleteBucket"
            ]
            "Resource": [
                "*"
            ]
        }
    ]
}'

ASSUME_POLICY_DOCUMENT='{
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
}'


account=$(aws sts get-caller-identity --output text --query 'Account')

echo "Creating policy"
aws iam create-policy --policy-name=$POLICY_NAME --policy-document="$HOUSEKEEPER_POLICY"
echo "Creating role"
aws iam create-role --role-name=$ROLE_NAME --assume-role-policy-document="$ASSUME_POLICY_DOCUMENT"
echo "Attaching policy to role"
aws iam attach-role-policy --role-name=$ROLE_NAME --policy-arn=arn:aws:iam::${account}:policy/$POLICY_NAME
