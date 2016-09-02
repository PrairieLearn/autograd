.PHONY: docker-image clean

docker-image:
	docker build -t prairielearn/autograd .

clean:
	rm -rf build
