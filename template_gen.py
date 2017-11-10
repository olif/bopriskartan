#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import argparse
import os
import sys
import yaml

from jinja2 import Environment, FileSystemLoader, select_autoescape


def main():
    description = ('Generates a google maps page with the heatmaps as custom '
    'overlays. The program takes a list of paths to the heatmaps as input to '
                   'stdin or they can be provided with the -f flag')
    parser = argparse.ArgumentParser(description=description)
    parser.add_argument('-c', '--conf', help='Path to config file',
                        required=True, dest='config_path')
    parser.add_argument('-t', '--template', help='Path to template file',
                        default='./', dest='template_path')
    parser.add_argument('-b', '--buckets', help='Path to bucket file',
                        default='bucket.json', dest='bucket_path')
    parser.add_argument('-f', '--files', dest='files',
                        help=('(optional) Comma seperated list of paths'
                              'to heatmap files'))

    args = parser.parse_args()

    file_paths = None
    if args.files is not None:
        files.paths = [s.strip() for s in args.files.split(',')]
    else:
        files = [s.strip() for s in sys.stdin.readlines()]
        file_paths = files

    config = None
    buckets = None

    with open(args.config_path, 'r') as config_file:
        config = yaml.load(config_file)

    with open(args.bucket_path, 'r')  as bucket_file:
        buckets = bucket_file.read()

    env = Environment(
        loader=FileSystemLoader(os.path.dirname(args.template_path)),
        autoescape=select_autoescape(['html']))

    config['buckets'] = buckets
    file_paths.sort()
    config['dates'] = file_paths
    template = env.get_template(os.path.basename(args.template_path))

    config_base_path = os.path.dirname(args.config_path)
    index_file_path = os.path.join(config_base_path, "index.html")
    print(index_file_path)
    with open(index_file_path, 'w') as index_file:
        index_file.write(template.render(config))

    print('A html file has been generated to: {0}'.format(index_file_path))

if __name__ == "__main__":
    main()
