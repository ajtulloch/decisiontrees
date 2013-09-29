all: test

test:
	python -m test.test
	./static/scripts/e2e-test.sh
	./static/scripts/test.sh

run:
	python -m run

fake:
	python -m test.fake_data
	
clean: 
	find . -iname *.pyc | xargs rm

lint:
	flake8 . --exclude=env,protobufs,shell.py

.PHONY: test

