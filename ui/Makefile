all: test

test:
	python test/test.py
	./static/scripts/e2e-test.sh
	./static/scripts/test.sh

run:
	python run.py

fake:
	python test/fake_data.py 
	
clean: 
	find . -iname *.pyc | xargs rm

.PHONY: test

