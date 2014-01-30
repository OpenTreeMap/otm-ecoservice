## Ecobenefits REST service based in Go

### Example Calculation

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
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__electricity.csv

The diameter ranges are in centimeters so we convert 20 in into 50.8
cm. Using those data we find the following range:

```
            38.10   53.34
BDS OTHER  189.2   189.2
```

In this case we're interpolating over a horizontal line so we get 189.2 kwh.

Doing the same for natural gas:
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__natural_gas.csv

```
            38.10   53.34
BDS OTHER  -81.4   -81.4
```

So the natural gas savings is computed to be -81.4 kbtus. That ends up
as -0.814 therms or -23.86 kwh

Continuing with stormwater:
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__hydro_interception.csv

```
            38.10   53.34
BDS OTHER    3.16    3.16
```

So we get 3.16 m^3 of water, which converts to 842 gal

Carbon calculations can be performed in the same manner. For storage:
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__co2_storage.csv

```
            38.10   53.34
BDS OTHER  569.6   569.6
```

569.6 kgs of CO2 stored converts to 1315 lbs.

CO2 avoided:
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__co2_avoided.csv

```
            38.10   53.34
BDS OTHER   55.7    55.7
```

55.7 kgs of CO2 avoided converts to 122.8 lbs

CO2 sequestered:
https://github.com/azavea/ecobenefits/blob/master/data/output__InlEmpCLM__co2_sequestered.csv

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


### Installation

TODO
