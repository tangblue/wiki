#!/usr/bin/env python

import yaml
import datetime

output_file = "example.yaml"
data = dict(
    a = datetime.datetime.now(),
    b = dict(
        c = 2,
        d = ['d', 'e']
    )
)

try:
    with open(output_file, 'r') as f:
       data = yaml.load(f)
except (IOError, yaml.YAMLError) as e:
    print(e)

print(data)

with open(output_file, 'w') as f:
    yaml.dump(data, f, default_flow_style=False)

with open(output_file, 'r') as f:
    try:
        print(yaml.load(f))
    except yaml.YAMLError as exc:
        print(exc)
