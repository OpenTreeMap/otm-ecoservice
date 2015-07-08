# ecoservice [![Build Status](https://travis-ci.org/OpenTreeMap/otm-ecoservice.svg?branch=master)](https://travis-ci.org/OpenTreeMap/otm-ecoservice)

REST services for calculating eco benefits for trees in an [OpenTreeMap](https://github.com/OpenTreeMap) database.

## Getting Started

In order make the process of developing against or building this project easier, we've wrapped everything inside of a [Vagrant](https://www.vagrantup.com/) project. In addition to Vagrant, Ansible is used to provision the Vagrant virtual machine. This ensures that all of the dependencies are in place.

First, ensure that Vagrant 1.5+ and Ansible 1.4.2+ are installed on your local workstation.

Next, start the Vagrant virtual machine:

```bash
$ vagrant up
```

You will see some Ansible output after the machine boots up. After that, SSH into the machine and execute the test suite:

```bash
$ vagrant ssh
vagrant@otm-ecoservice:~$ cd src/github.com/OpenTreeMap/otm-ecoservice/
vagrant@otm-ecoservice:~/src/github.com/OpenTreeMap/otm-ecoservice$ make test
godep go test eco/*
ok      command-line-arguments  0.494s
```

If you want to build a release, use the `release` target:

```bash
vagrant@otm-ecoservice:~/src/github.com/OpenTreeMap/otm-ecoservice$ make release
vagrant@otm-ecoservice:~/src/github.com/OpenTreeMap/otm-ecoservice$ exit
$ ls -l *.tar.gz
-rw-r--r--  1 hcastro  staff  2714304 Sep  9 15:41 ecoservice.tar.gz
```

## Running the ``ecoservice``

The ``ecobenefits`` executable must be run with a configuration file
containing three sections:

* ``database`` - PostgreSQL database connection information
* ``server`` - The host address and port on which the service will listen
* ``data`` - The absolute path to the data directory (with a trailing
  slash)

A template of this config file is provided in ``config.gcfg.template``.

Once a config file has been created, the ``ecobenefits`` service can be launched with:

```bash
$ /path/to/ecobenefits --configpath=/path/to/config.gcfg
```

## Example Calculation

20 inch Common Fig

Identify the OTM2/USDA code for the given species:
FICA

Determine the region the tree is in. For our example
we'll use Inland Empire (InlEmpCLM).

Determine the itree code using the master species list
and the selected region.
BDS OTHER

Calculations then are determined by linear interpolation
from the root resource sheets.

For the electricity case we use this csv:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__electricity.csv

The diameter ranges are in centimeters so we convert 20 in into 50.8
cm. Using those data we find the following range:

```
            38.10   53.34
BDS OTHER  189.2   189.2
```

In this case we're interpolating over a horizontal line so we get 189.2 kwh.

Doing the same for natural gas:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__natural_gas.csv

```
            38.10   53.34
BDS OTHER  -81.4   -81.4
```

So the natural gas savings is computed to be -81.4 kbtus. That ends up
as -0.814 therms or -23.86 kwh

Continuing with stormwater:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__hydro_interception.csv

```
            38.10   53.34
BDS OTHER    3.16    3.16
```

So we get 3.16 m^3 of water, which converts to 842 gal

Carbon calculations can be performed in the same manner. For storage:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__co2_storage.csv

```
            38.10   53.34
BDS OTHER  569.6   569.6
```

569.6 kgs of CO2 stored converts to 1315 lbs.

CO2 avoided:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__co2_avoided.csv

```
            38.10   53.34
BDS OTHER   55.7    55.7
```

55.7 kgs of CO2 avoided converts to 122.8 lbs

CO2 sequestered:
https://github.com/OpenTreeMap/otm-ecoservice/blob/master/data/output__InlEmpCLM__co2_sequestered.csv

```
            38.10   53.34
BDS OTHER   24.6     0.0
```

Here we have to actually calculate:

```
m = dx/dy = -24.6 / 15.24 = -1.614
y0 = y - mx = 0.0 - 53.34 * -1.614 = 86.1
y = mx + b = -1.614 * 50.8 + 86.1 = 4.1 kgs = 9.06 lbs
```

We can test this particular case by running an eco benefits server and
using the curl utility:

```
$ curl "localhost:13000/eco.json?otmcode=FICA&diameter=20&region=InlEmpCLM"

{
  "Benefits": {
    "aq_nox_avoided": 0.1102,
    "aq_nox_dep": 0.119,
    "aq_ozone_dep": 0.35,
    "aq_pm10_avoided": 0.0273,
    "aq_pm10_dep": 0.185,
    "aq_sox_avoided": 0.2183,
    "aq_sox_dep": 0.016,
    "aq_voc_avoided": 0.0273,
    "bvoc": 0,
    "co2_avoided": 55.7,
    "co2_sequestered": 4.1000000000000085,
    "co2_storage": 569.6,
    "electricity": 189.2,
    "hydro_interception": 3.16,
    "natural_gas": -81.4
  }
}
```

### Terminology

#### Factors
What i-Tree calls 'benefit categories', we refer to as 'factors' in our source. These are distinct ways in which environmental influence can be quantified for trees. Examples include 'CO2 avoided' and 'electricity (saved)'.
