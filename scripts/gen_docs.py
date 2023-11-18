import os
import subprocess
import fnmatch

PKGS=["blueprint", "plugins", "runtime"]

def main():
    '''
    Script assumes that you are running it from the root directory.
    Expected Usage: python3 scripts/gen_docs.py
    '''
    print("Removing old auto-generated documentation")
    rc = subprocess.call("./scripts/cleanup_autodocs.sh")

    for pkg in PKGS:
        for root, dirs, files in os.walk(pkg):
            filtered_files = fnmatch.filter(files, "*.go")
            # Current directory has some go files so we can auto-generate documentation
            if len(filtered_files) != 0:
                print("Generating documentation for ", root)
                subprocess.run(["./scripts/gen_docs.sh", root])

if __name__ == '__main__':
    main()
