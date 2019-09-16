// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: BSD-2-Clause

// Package cloudsweeper is where all the 'sweeping logic is defined. Such
// as what to notify about and what to clean up. For using Cloudsweeper in
// larger organizations, the Organization structure was implemeted.
package cloudsweeper

import (
	"encoding/json"
	"fmt"

	"github.com/agaridata/cloudsweeper/cloud"
)

// Organization represents the employees, their departments, and their managers
// within an organization. This structure was set up for an org wherein all
// employees have their own cloud accounts, and are aggregated under a single
// payer account. In the case you have only a single account, this will be
// superfluous.
type Organization struct {
	Managers    Employees   `json:"-"`
	ManagerIDs  []managerID `json:"managers"`
	Departments Departments `json:"departments"`
	Employees   Employees   `json:"employees"`

	managerMapping    map[string]*Employee
	departmentMapping map[string]*Department
	employeeMapping   map[string]*Employee
	managerEmployees  map[string]Employees
}

type managerID struct {
	ID string `json:"username"`
}

// Department represents a department in your org
type Department struct {
	Number int    `json:"number"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

// Departments is a list of Department
type Departments []*Department

// Employee represents an employee, which
// belong to a department and has a manager. An employee can
// also have multiple accounts and projects associated with
// them in AWS and GCP. "Disabled" employees are employees
// who should no longer be regarded as active in the company
type Employee struct {
	Username     string      `json:"username"`
	RealName     string      `json:"real_name"`
	ManagerID    string      `json:"manager"`
	Manager      *Employee   `json:"-"`
	DepartmentID string      `json:"department"`
	Department   *Department `json:"-"`
	Disabled     bool        `json:"disabled,omitempty"`
	AWSAccounts  AWSAccounts `json:"aws_accounts"`
	GCPProjects  GCPProjects `json:"gcp_projects"`
}

// Employees is a list of Employee
type Employees []*Employee

// AWSAccount represents an account in AWS. An account
// can have automatic cleanup enabled, indiacated by
// the CloudsweeperEnabled attribute.
type AWSAccount struct {
	ID                  string `json:"id"`
	CloudsweeperEnabled bool   `json:"cloudsweeper_enabled,omitempty"`
}

// AWSAccounts is a list of AWSAccount
type AWSAccounts []*AWSAccount

// GCPProject represents a project in GPC. A project
// can have automatic cleanup enabled, indiacated by
// the CloudsweeperEnabled attribute.
type GCPProject struct {
	ID                  string `json:"id"`
	CloudsweeperEnabled bool   `json:"cloudsweeper_enabled,omitempty"`
}

// GCPProjects is a list of GCPProject
type GCPProjects []*GCPProject

// InitOrganization initializes an organisation from raw data,
// e.g. the contents of a JSON file.
func InitOrganization(orgData []byte) (*Organization, error) {
	org := new(Organization)
	err := json.Unmarshal(orgData, org)
	if err != nil {
		return nil, err
	}
	org.departmentMapping = make(map[string]*Department, len(org.Departments))
	for i := range org.Departments {
		org.departmentMapping[org.Departments[i].ID] = org.Departments[i]
	}
	// First initalize all employees
	org.employeeMapping = make(map[string]*Employee, len(org.Employees))
	for i := range org.Employees {
		org.employeeMapping[org.Employees[i].Username] = org.Employees[i]
		if department, exist := org.departmentMapping[org.Employees[i].DepartmentID]; exist {
			org.Employees[i].Department = department
		} else {
			// TODO: Fail if employee's department doesn't exist
		}
	}
	// Then map the employees' managers
	org.managerMapping = make(map[string]*Employee, len(org.Managers))
	org.Managers = Employees{}
	for i := range org.ManagerIDs {
		if manager, exist := org.employeeMapping[org.ManagerIDs[i].ID]; exist {
			org.managerMapping[org.ManagerIDs[i].ID] = manager
		} else {
			// A manager doesn't have an record in the employee list
			return nil, fmt.Errorf("Manager %s is not in the list of employees", org.ManagerIDs[i])
		}
		org.Managers = append(org.Managers, org.employeeMapping[org.ManagerIDs[i].ID])
	}
	org.managerEmployees = make(map[string]Employees, len(org.Managers))
	for i := range org.Employees {
		if manager, exist := org.managerMapping[org.Employees[i].ManagerID]; exist {
			org.Employees[i].Manager = manager
			org.managerEmployees[manager.Username] = append(org.managerEmployees[manager.Username], org.Employees[i])
		} else {
			// TODO: Fail if employee's manager doesn't exist
		}
	}
	return org, nil
}

// EmployeesForManager gets all the employees who has the
// specifed manager as their manager.
func (org *Organization) EmployeesForManager(manager *Employee) (Employees, error) {
	if _, isManager := org.managerMapping[manager.Username]; !isManager {
		return nil, fmt.Errorf("%s is not a manager", manager.Username)
	}
	if employees, exist := org.managerEmployees[manager.Username]; exist {
		return employees, nil
	}
	// Manager has no employees
	return Employees{}, nil
}

// EnabledAccounts will return a list of all cloudsweeper enabled accounts
// in the specified CSP
func (org *Organization) EnabledAccounts(csp cloud.CSP) []string {
	accounts := []string{}
	for _, employee := range org.Employees {
		switch csp {
		case cloud.AWS:
			for _, account := range employee.AWSAccounts {
				if account.CloudsweeperEnabled {
					accounts = append(accounts, account.ID)
				}
			}
		case cloud.GCP:
			for _, project := range employee.GCPProjects {
				if project.CloudsweeperEnabled {
					accounts = append(accounts, project.ID)
				}
			}
		}
	}
	return accounts
}

// AccountToUserMapping is a helper method that maps accounts to their owners
// username. This is useful for sending out emails to the owner of an account.
func (org *Organization) AccountToUserMapping(csp cloud.CSP) map[string]string {
	result := make(map[string]string)
	for _, employee := range org.Employees {
		switch csp {
		case cloud.AWS:
			for _, account := range employee.AWSAccounts {
				result[account.ID] = employee.Username
			}
		case cloud.GCP:
			for _, project := range employee.GCPProjects {
				result[project.ID] = employee.Username
			}
		}
	}
	return result
}

// UsernameToEmployeeMapping is a helper method that returns a map of username to Employee struct.
func (org *Organization) UsernameToEmployeeMapping() map[string]*Employee {
	return org.employeeMapping
}
