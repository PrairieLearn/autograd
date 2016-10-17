#!/bin/bash
kubectl patch deployment autograd-cs225 -p "{\"spec\":{\"template\":{\"spec\":{\"containers\":[{\"name\":\"autograd-cs225\",\"env\":[{\"name\":\"RESTART_\",\"value\":\"$(date -uIseconds)\"}]}]}}}}"
