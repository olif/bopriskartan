#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse
import csv
import os
import sys
import yaml

from datetime import datetime

import booli


class Scraper(object):

    def __init__(self, config, user, key):
        self.config = config
        self.user = user
        self.key = key

    def extract_fields(self, o):
        """Extract relevant fields from a booli sold object"""
        booli_id = o['booliId']
        living_area = o['livingArea']
        price = o['soldPrice']
        location = o['location']
        position = location['position']
        lat = position['latitude']
        lng = position['longitude']
        sold_date = o['soldDate']
        return {'id': booli_id, 'livingArea': living_area, 'price': price,
                'lat': lat, 'lng': lng, 'soldDate': sold_date}

    def run(self, from_date, to_date, object_type):
        bbox = '{lat_lo},{lng_lo},{lat_hi},{lng_hi}'.format(**self.config)
        start_date_str = datetime.strftime(from_date, "%Y%m%d")
        end_date_str = datetime.strftime(to_date, "%Y%m%d")
        queryDict = {"bbox": bbox, "minSoldDate": start_date_str,
                     "maxSoldDate": end_date_str}

        if object_type is not None:
            queryDict['objectType'] = object_type

        headers = ['id', 'livingArea', 'price', 'lat', 'lng', 'soldDate']
        writer = csv.DictWriter(sys.stdout, headers)
        writer.writeheader()

        booli_client = booli.BooliClient(self.user, self.key)
        for page in booli_client.get_sold_objects(queryDict):
            for sold in page.payload:
                try:
                    o = self.extract_fields(sold)
                    writer.writerow(o)
                except KeyError as err:
                    sys.stderr.write("Unexpected object, skipping row\n")
                except:
                    sys.stderr.write("Unexpected error, skipping row\n")


def main():
    booli_user_name = os.getenv("BOOLI_USERNAME")
    booli_key = os.getenv("BOOLI_KEY")

    parser = argparse.ArgumentParser(description='Booli API scraper for sold '
    'objects')

    parser.add_argument('-c', '--conf', help='Path to config file',
                        required=True, dest='config_path')
    parser.add_argument('-f', '--from_date', help='Date to query objects from',
                        metavar='2017-01-01', required=True, dest='from_date')
    parser.add_argument('-t', '--to_date', help='Date to query objects to',
                        metavar='2017-01-07', required=True, dest='to_date')
    parser.add_argument('-o', '--object_type',
                        help='Type of object to query (optional)',
                        metavar=('{villa, lägenhet, gård, tomt-mark,'
                                 'fritidshus, parhus,radhus,kedjehus}'),
                        required=False, dest='object_type')

    args = parser.parse_args()

    if booli_user_name == None or booli_key == None:
        raise ValueError('Booli credentials not set in env')

    start_date = datetime.strptime(args.from_date, "%Y-%m-%d")
    end_date = datetime.strptime(args.to_date, "%Y-%m-%d")

    config = None
    with open(args.config_path, 'r') as config_file:
        config = yaml.load(config_file)

    scraper = Scraper(config, booli_user_name, booli_key)
    scraper.run(start_date, end_date, args.object_type)

if __name__ == '__main__':
    main()
