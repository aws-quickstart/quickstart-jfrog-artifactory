.PHONY: help run


help:
    @echo   "make test  : executes taskcat"


test:
	cd .. && \
	taskcat -c theflash/ci/config.yml

