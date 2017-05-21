#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import random
import string
import time
import requests
import urllib.request as urllib2
from hashlib import sha1
from urllib.parse import urlencode


def html_decode(s):
    return s.replace("&lt;", "<").replace("&gt;", ">").replace("&amp;", "&")


def urlify_value(value):
    if isinstance(value, int):
        return str(int(value))
    elif isinstance(value, list):
        return ",".join(urlify_value(x) for x in value)
    return str(value)


def smart_urlencode(params):
    return urlencode(dict((key, urlify_value(value))
                          for key, value in params.items()))


class BooliClient(object):
    base_url = "http://api.booli.se"

    def __init__(self, caller_id, key):
        self.caller_id = caller_id
        self.key = key

    def get_sold_objects(self, params):
        last = False
        limit = 500
        offset = 0
        while not last:
            response = self._get_sold_object(limit, offset, params)
            yield response
            offset += limit
            last = response.offset > response.totalCount

    def _get_sold_object(self, limit, offset, params):
        url = self.base_url + "/sold"
        params.update(limit=limit, offset=offset)
        params.update(self._get_auth_params())
        response = requests.get(url, params)
        return BooliResponse.from_dict(response.json())

    def _get_auth_params(self):
        timestamp = str(time.time()).split('.')[0]
        unique = "".join(random.choice(string.ascii_letters + string.digits)
                         for _ in range(16))
        line = (self.caller_id + timestamp + self.key + unique).encode('utf-8')
        hash = sha1(line).hexdigest()
        params = {}
        params.update(callerId=self.caller_id, time=timestamp,
                      unique=unique, hash=hash, format="json")
        return params


class BooliAPI(object):
    base_url = "http://api.booli.se/sold"

    def __init__(self, caller_id, key):
        self.caller_id = caller_id
        self.key = key

    def search(self, area="", **params):
        url = self._build_url(area, params)
        print(url)
        response = urllib2.urlopen(url)
        data = json.load(response)
        bs = BooliResponse.from_dict(data)
        print(bs.totalCount)
        return bs

    def _build_url(self, area, params):
        """Return a complete API request URL for the given search
            parameters, including the required authentication bits."""
        timestamp = str(time.time()).split('.')[0]
        unique = "".join(random.choice(string.ascii_letters + string.digits)
                         for _ in range(16))
        line = (self.caller_id + timestamp + self.key + unique).encode('utf-8')
        hash = sha1(line).hexdigest()
        params.update(q=area, callerId=self.caller_id, time=timestamp,
                      unique=unique, hash=hash, format="json")
        return self.base_url + "?" + smart_urlencode(params)


class BooliResponse(object):

    def __init__(self, totalCount, count, limit, offset, payload):
        self.totalCount = totalCount
        self.count = count
        self.limit = limit
        self.offset = offset
        self.payload = payload

    def from_dict(o):
        totalCount = o['totalCount']
        count = o['count']
        limit = o['limit']
        offset = o['offset']
        payload = o['sold']
        return BooliResponse(totalCount, count, limit, offset, payload)
