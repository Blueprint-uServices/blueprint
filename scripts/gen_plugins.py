'''
This script generates index files for all the available plugins for Blueprint's website.
The generated index files are placed at https://github.com/Blueprint-uServices/Blueprint-uServices.github.io/tree/main/_documents
which then get listed on the Blueprint's website at https://blueprint-uservices.github.io/plugins/.
'''
import os
import sys
from datetime import date
import re

BASE_URL="https://github.com/Blueprint-uServices/blueprint/tree/main/plugins"

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 scripts/gen_plugins.py <path/to/out/folder")
        sys.exit(1)
    out_folder = sys.argv[1]
    plugins_path = "./plugins"
    all_plugins = []
    for root, dirs, files in os.walk(plugins_path):
        if root != plugins_path:
            continue
        for d in dirs:
            all_plugins += [d]
    all_plugins = sorted(all_plugins)

    today = date.today()
    for p in all_plugins:

        # Read package description
        description = ""
        with open(os.path.join(plugins_path, p, 'README.md'), 'r') as inf:
            all = inf.read()
            format = r'([P|p]ackage ' + p + r'.*\n*.*)'
            matches = re.findall(format, all)
            if len(matches) != 0:
                description = matches[0].replace("\n", "").replace(":", "-")
            else:
                print("Package description not found for ", p)
        path = BASE_URL + "/" + p
        name = p
        with open(os.path.join(out_folder, name + ".md"), 'w') as outf:
            s = "---\n"
            s += f"title: {name}\n"
            s += f"target: {path}\n"
            s += f"date: {today}\n"
            s += f"description: {description}\n"
            s += "---\n"
            outf.write(s)

if __name__ == '__main__':
    main()