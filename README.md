# Bopriskartan

Bopriskartan is a tool for generating heatmaps, visualizing the price 
development of housing. It takes a csv file as input and generates a png file 
as output. This png can then be used as a custom layer on for instance, google 
maps.

A simple scraper script is included which fetches sold prices from 
[Booli](http://www.booli.se):s Api and prints the to stdout.

## Getting started
In order to use the tool, [Go](https://golang.org/) must be installed on your 
machine.

Clone the repository to $GOPATH/src.


## Building
The project uses [dep](https://github.com/golang/dep) as dependency management
tool. In order to build the project, go/dep must first be installed.

    $> go get -u github.com/golang/dep/cmd/dep 

The project can then be built by the included Makefile.

    $> make
    
or it can be compiled directly with go build from the src directory.

    $> go get
    $> go build heatmap.go main.go -o heatmap
    
## Creating a heatmap
In order to create a heatmap we need one or more data files containing prices
for sold housing objects and a configuration file which tells both the scraper
and the heatmap generator which area the prices are covering.


### Configuration file
The configuration must be written in yaml format and contain the following
fields:

    ---
    name: stockholm
    heatmap_width: 1000
    heatmap_height: 1000
    heatmap_step: 5000
    min_price_limit: 20000
    max_price_limit: 120000
    lat_lo: 59.205634
    lng_lo: 17.846684
    lat_hi: 59.439306
    lng_hi: 18.266219
    lat_center: 59.324818
    lng_center: 18.072342
    zoom: 13
    ...
    

| Name                   | Description                                                                                                                   |
|------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| name                   | Title of the html page                                                                                                        |
| heatmap_width          | The width of the generated png                                                                                                |
| heatmap_height         | The height of the generated png                                                                                               |
| heatmap_step           | The step in price between max_price_limit and min_price_limit, determines how many color buckets will be generated in between |
| min_price_limit        | First color bucket                                                                                                            |
| max_price_limit        | Last color bucket                                                                                                             |
| lat_lo, lat_hi         | Latitude values                                                                                                               |
| lng_lo, lng_hi         | Longtitude values                                                                                                             |
| lat_center, lng_center | The center point on the generated map                                                                                         |
| zoom                   | The zoom level of the generated map                                                                                           |

lat_lo, lng_lo, lat_hi and lng_hi should together create a rectangular box of
a geographic area. (lat_hi, lng_hi) are the northeast corner and 
(lat_lo, lng_lo) are south west corner.

### Scraping data
A small python (3) script is included in the repo which loads data from the 
[Booli](http://www.booli.se) Api and prints the data to stdout in correct 
format. 


    $>./scraper.py -h
    usage: scraper.py [-h] -c CONFIG_PATH -f 2017-01-01 -t 2017-01-07
                      [-o {villa, lägenhet, gård, tomt-mark,fritidshus, parhus,radhus,kedjehus}]

    Booli API scraper for sold objects

    optional arguments:
      -h, --help            show this help message and exit
      -c CONFIG_PATH, --conf CONFIG_PATH
                            Path to config file
      -f 2017-01-01, --from_date 2017-01-01
                            Date to query objects from
      -t 2017-01-07, --to_date 2017-01-07
                            Date to query objects to
      -o {villa, lägenhet, gård, tomt-mark,fritidshus, parhus,radhus,kedjehus}, --object_type {villa, lägenhet, gård, tomt-mark,fritidshus, parhus,radhus,kedjehus}
                            Type of object to query (optional)

In order to use the script, you need to have a booli API account.
The username and key must then be set as environment variables.

    $> export BOOLI_USERNAME={your username}
    $> export BOOLI_KEY={your key}

Then the scraper can be used to pull price data from the API.

    $> ./scraper.py -c {path/to/config.yml} -f 2017-01-01 -t 2017-01-31 -o 'lägenhet' > data.csv

The data file has the following structure:

    $> head data.csv
    id,rooms,livingArea,soldPrice,lat,lng
    2267369,1,24,1600000,59.2214,17.9462
    2283986,3,68,4450000,59.31301432,18.05819908
    2281051,2,60,2520000,59.24037365,18.09433291
    2243121,3,66,5525000,59.34588416,18.05591002
    2276347,4,75,3600000,59.27839987,18.1310667
    
    
### Generating a heatmap from the data file
The heatmaps are generated with the heatmap_gen binary.

    $> ./heatmap -h
    Usage of ./heatmap:
      -b	Prints bucket list to output
      -c string
            Path to config file (default "config.yml")
      -d string
            Path to input data
      -o string
            Path to output (default "/.")


When you have a csv file containing price data, a heatmap can be generated:

    $> cat data.csv | heatmap_gen -c {path/to/config.yml} -o heatmap.png

Also, the color buckets can be generated to a file as a json object.

    $> ./heatmap_gen -c {path/to/config.yml} -o color_buckets.json -b
    
    
### Creating the map
Now when we have the heatmap image we can add it as a custom layer ontop of a 
map. A python (3) scripts is included which makes use of a html template in 
jinja format. The generated html file will be placed in the same directory as 
the configuration file.

    $> ./template_gen.py -h
    usage: template_gen.py [-h] -c CONFIG_PATH [-t TEMPLATE_PATH] [-b BUCKET_PATH]
                          [-f FILES]

    Generates a google maps page with the heatmaps as custom overlays. The program
    takes a list of paths to the heatmaps as input to stdin or they can be
    provided with the -f flag

    optional arguments:
      -h, --help            show this help message and exit
      -c CONFIG_PATH, --conf CONFIG_PATH
                            Path to config file
      -t TEMPLATE_PATH, --template TEMPLATE_PATH
                            Path to template file
      -b BUCKET_PATH, --buckets BUCKET_PATH
                            Path to bucket file
      -f FILES, --files FILES
                            (optional) Comma seperated list of pathsto heatmap
                            files 

Given a configuration file, a heatmap, a bucket.json and a template (which is
included) a html file can be generated.

    $> find . -name "*.png" | ./template_gen -c /{path/to/config.yml} -t template.html.j2 -b color_buckets.json
