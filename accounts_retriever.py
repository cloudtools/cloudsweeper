"""
This simple python script is used to get the list of accounts
from the metavisor repository, and then stores them in a file
so tha they can be used by the Go program.
"""
import argparse
import json

from csp_utils.aws.accounts_config import (
    BRKT_EMPLOYEE_ACCOUNTS
)


def main(output):
    """
    This script will get the accounts constant from the MV repo.
    It then creates a list of owner structures, such as:
    [
        {
            "name": "qa",
            "id": "475063612724"
        }
    ]
    This list is saved in a file at the specified location
    """
    account_list = []
    for name, account in BRKT_EMPLOYEE_ACCOUNTS.iteritems():
        owner = {
            'name': name,
            'id': account
        }
        account_list.append(owner)
    accounts_json = json.dumps(account_list, indent=4, sort_keys=True)
    output_file = open(output, 'w')
    output_file.write(accounts_json)
    output_file.close()


if __name__ == '__main__':
    PARSER = argparse.ArgumentParser()
    PARSER.add_argument('--output', default='./aws_accounts.json',
                        help='Where to store accounts')
    ARGS = PARSER.parse_args()
    main(ARGS.output)
