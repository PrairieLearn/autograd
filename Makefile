.PHONY: docker-image clean

docker-image:
	docker build -t autograd .

clean:
	rm -rf build
