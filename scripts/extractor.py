#!/usr/bin/env python

# Extract species and eco benefit factor data from the Excel workbooks
# included in the i-Tree Streets installation

import argparse
import csv
import os
import subprocess

from bs4 import BeautifulSoup

XLS2CSV_EXEC = 'xls2csv'


def _assert_has_xls_converter():
    try:
        subprocess.check_call(['which', XLS2CSV_EXEC])
    except subprocess.CalledProcesError:
        print 'You need to have a supported version of "xls2csv" installed\n'\
            'Install it on ubuntu with "apt-get install catdoc"'
        exit(1)


def is_valid_code(code):
    return (code and
            len(code.strip()) and
            code.strip() != '"SpeciesCode"')


def is_valid_name(name):
    """ This is used for excel conversion weirdness """
    try:
        float(name)
        return False
    except:
        return True


def parse_path(soup, path):
    ids = {}
    for tag in soup.find('p').find_all('a'):
        ids[tag['href'][1:]] = tag.get_text().lower().replace(' ', '_')

    for table in soup.find_all('table'):
        ref_link = table.find_previous_sibling('a')
        category = ids[ref_link['name']]
        file_to_write = "%s__%s.csv" % (path, category)

        writer = file(file_to_write, 'w')

        for row in table.find_all('tr'):
            cells = [
                cell.get_text().replace(',', '') for cell in row.find_all('td')
            ]
            writer.write(','.join(cells) + "\n")


def extract_data(output_dir, resource_dir):
    for root, dirs, files in os.walk(resource_dir):
        if 'ResourceUnit.html' in files:
            html_path = os.path.join(root, 'ResourceUnit.html')
            soup = BeautifulSoup(file(html_path).read(), 'html.parser')
            parse_path(soup,
                       os.path.join(output_dir,
                                    'output__%s' % os.path.split(root)[1]))


def extract_species(output_dir, resource_dir):
    _assert_has_xls_converter()

    header = ['SpeciesCode', 'ScientificName', 'CommonName', 'Tree Type',
              'SppValueAssignment', 'Species Rating (%)',
              'Basic Price ($/sq in)', 'Palm Trunk Cost($/ft)',
              'Replacement Cost ($)', 'TAr (sq Inches)', 'region']

    file_path = os.path.join(output_dir, 'species_master_list.csv')
    output_file = file(file_path, 'w')

    writer = csv.DictWriter(output_file, header)
    writer.writeheader()

    devnull = open('/dev/null', 'w')

    for root, dirs, files in os.walk(resource_dir):
        if 'SpeciesCode.xls' in files:
            region_code = os.path.split(root)[1]

            species_path = os.path.join(root, 'SpeciesCode.xls')
            p = subprocess.Popen([XLS2CSV_EXEC, species_path],
                                 stdout=subprocess.PIPE,
                                 stderr=devnull)
            csvdata, err = p.communicate()

            reader = csv.DictReader(csvdata.split('\n'))

            for row_dict in reader:
                code = row_dict['SpeciesCode']
                name = row_dict['ScientificName']

                if is_valid_code(code) and is_valid_name(name):
                    row_dict['region'] = region_code
                    writer.writerow(row_dict)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('action', help='Action should be "extract_species" or'
                                       '"extract_values"')
    parser.add_argument('-r', '--resource-dir', help='resource unit directory')
    parser.add_argument('-d', '--output-dir', help='output directory')
    args = parser.parse_args()

    action = args.action
    output_dir = args.output_dir or ''
    resource_dir = args.resource_dir or 'ResourceUnit'

    if action != 'extract_species' and action != 'extract_values':
        parser.print_help()
        exit(1)

    if not os.path.exists(resource_dir):
        print 'Error: Could not find a valid resource directory at {}. \n'\
            'Specify one with "-r"'.format(resource_dir)
        exit(1)

    if action == 'extract_species':
        extract_species(output_dir, resource_dir)
    elif action == 'extract_values':
        extract_data(output_dir, resource_dir)


if __name__ == '__main__':
    main()
