# Cloudsweeper Instructions

## Content

- [Overview](#Overview)
- [Account Setup](#account-setup)
    - [Master Setup](#master-setup)
    - [Slave Setup](#slave-setup)
- [Organization definition](#organization-definition)
- [Configuration](#configuration)
- [Building](#building)
- [Functionality](#functionality)

## Overview
What is Cloudsweeper? Cloudsweeper is a tool that, once properly setup, can monitor and take action in multiple cloud accounts (supporting both AWS and GCP). The problem it solved when first created was to make sure that all cloud accounts in the compnay were used as effectively as possible, not leaving any unused resources sticking around.

At Bracket Computing, Inc. (where it was originally created), Cloudsweeper was ran daily to perform different operations that monitored and cleaned up all employees' accounts.

## Account Setup
There are two parts to properly setting up Cloudsweeper; setting up the _master_, and setting up all employee/slave accounts.

### Master setup
The _master_ is the machine/server where Cloudsweeper will run from. This will be responsible for _assuming_ into the slave accounts to perform monitoring and/or cleanup.

Typically, the master will always run from the same machine or system — perhaps form a Jenkins instance? It would be recommended to setup the master to run at some defined interval, performing the actions you want.

For AWS, you should have the master always run from the same account, using the same role. This is important to let the slave accounts only need to allow role assumption from a single ARN. Start by deciding which account to run Cloudsweeper with, note its account ID. Then create an IAM user in this account that can assume into other accounts, note its name. When you have done this, you need the ARN, which will look like:

```
arn:aws:iam::<account ID>:user/<user name>
```

Let's call this the _master ARN_.

### Slave setup
The slaves are the accounts that Cloudsweeper will monitor. These need to be setup so that the master can access them.

For AWS, the easiest way to setup a slave account is to run either the built-in Cloudsweeper setup command or to run the `aws_setup.sh` script. Both of these methods assumes that the machine they are being run on are setup with AWS (i.e. having set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`). This will setup an IAM role in the account with the name `Cloudsweeper` that has an IAM policy with the name `CloudsweeperPolicy` — these names are important and can not be modified.

When the `Cloudsweeper` role is being created, an assume policy is attached to it which allows the _master ARN_ (see above) to assume into the role. If using any of the two automatic setup methods, make sure to first configure them to use your unique master ARN. If using the `aws_setup.sh` script, you must set the environment variable `CS_MASTER_ARN`. If using the built-in setup command you must either specify the flag `--aws-master-arn` when running the command or add the ARN to the `config.conf` file.

This setup must be done for all accounts that should be monitored by Cloudsweeper.

## Organization definition
Cloudsweeper uses a central organization definition file (see `example-org.json` for an example) to figure out which slaves/employees it should monitor. Every employee can own multiple different AWS and GCP accounts, and can have a manager assigned to them.

- `managers` is a list of all managers
- `departments` is a list of all departments
- `employees` is a list of all employees

All managers should also be definied in the list of employees with the same username. The username should preferably match with the person's email alias, as this will be used by Cloudsweeper to send out mail (it should just be the alias, i.e. the part before the `@`, as the domain part is configured). To enable cloudsweeper in an employee's account, it's important to specify `cloudsweeper_enabled: true`, as it defaults to `false` otherwise.

**NOTE:** Employees obviously don't need to be actual employees, they can be anything. An _employee_ could be the Production account for example, and another could be Stage.

## Configuration
In order for Cloudsweeper to work properly, besides the previously mentioned setup, it needs to be configured. The recommended way to configure Cloudsweeper is to use the `config.conf` file, however, all configuration can also be made through command line flags. Flags take precedence over the `config.conf` file, so it can be used to override anything specified in that file. The flags themselves can be discovered by either running `./cloudsweeper --help` or by looking in the `cmd/cloudsweeper/main.go` file.

The `config.conf` file contains descriptions of all configuration options.

## Building
Cloudsweeper was built using Go 1.10. In order to complile it, you need to either install Go or Docker. For building with Go, simply run:
```
$ go get ./...
$ go build -o cs cmd/cloudsweeper/*.go
```
Which will create a binary called `cs` which you can then execute (e.g `./cs setup`).

If using Docker, the easiest way to build is by using the make target `make build`, which will build a container `cloudsweeper:latest`. 

**IMPORTANT:** If building with Docker, modify the `Dockerfile` to use your own organization JSON file (it defaults to `example-org.json`). This could also reference a remote file in e.g. an S3 bucket.

## Functionality
The best way to explore what Cloudsweeper can do is to look at the source code. It is however divided into some different parts. A good way to start exploring is to look at the `notify` and the `cleanup` packages within the `cloudsweeper`. For example, the `cleanup` command will run the `PerformCleanup` function in the `cleanup` package. This function in turn delegates to other functions, and this would be an ideal place for you to add more things to be clean up if you wanted to extend Cloudsweeper.

All notification and cleanup functions leverages a filtering system which makes it really easy to first filter out the resources you wanna operate on and then perform some action (e.g. clean them up) — you could draw parallels to map-reduce.