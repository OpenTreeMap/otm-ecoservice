# Scripts

### extractor.py

Extract species and eco benefit factor data from the Excel workbooks included in the i-Tree Streets installation.

## Getting the i-Tree Streets data

- Download the i-Tree installer from
  https://www.itreetools.org/tools.php (Windows only, requires free
  account registration)
- Install the application. You may choose the custom option and only
  install Streets.
- Copy the data directory from `C:\Program
  Files(x86)\i-Tree\STREETS\Data` to `{ecoservice
  repo}/scripts/local/data`.

## Extracting the eco benefit factor CSVs

### Setting up the tools and data

- Install LibreOffice https://www.libreoffice.org/.
- Open the `ResourceUnit.xls` file for each i-Tree region, found at
  `{ecoservice repo}/scripts/local/data/ResourceUnit/{region code}/ResourceUnit.xls`
    - Cancel any prompts about macros or updating file links.
- Export each regional workbook as HTML.
    - Choose `File > Save as...`.
    - Choose `HTML Document (Calc)` as the file type.
    - Make sure the file is named `ResourceUnit.html` and is saved in
      the same directory as the .xls file.
- `vagrant ssh app` and then run `sudo /opt/ecoservice/scripts/setup.sh` to
  install the extraction dependencies.

### Extract the data

- On the app VM run the following:
    - `cd /opt/ecoservice/scripts`
    - `mkdir -p local/output`
    - `./extractor.py -r local/data/ResourceUnit -d local/output extract_values`
- On the host, confirm that the `{ecoservice repo}/scripts/local/output` folder
  contains a set of CSV files for each i-Tree region directory in the
  `{ecoservice repo}/scripts/local/data/ResourceUnit/` directory.
