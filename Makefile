# Copyright 2016 tsuru-admin authors. All rights reserved.
# Use of this source code is governed by a BSD-style
#  license that can be found in the LICENSE file.

doc: docs-clean
	@python docs/source/exts/tsuru_cmd.py && cd docs && make html SPHINXOPTS="-N -W"

docs: doc

docs-clean:
	@rm -rf ./docs/build

release:
	@if [ ! $(version) ]; then \
		echo "version parameter is required... use: make release version=<value>"; \
		exit 1; \
	fi

	@echo "Releasing tsuru-admin $(version) version."

	@echo "Replacing version string."
	@sed -i "" "s/version = \".*\"/version = \"$(version)\"/g" main.go

	@git add main.go
	@git commit -m "bump to $(version)"

	@echo "Creating $(version) tag."
	@git tag $(version)

	@git push --tags
	@git push origin master

	@echo "$(version) released!"
