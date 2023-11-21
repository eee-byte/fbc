#!/bin/bash

    for (( i=0; ; ))
    do
        $(pwd)/fbchaincli tx send node1 0xB526728759eD94Df2b1DDA90Ad9137Ee2236048d 1fibo --fees 1fibo -b block -y --home ./chain-data/node1/fbchaincli/ --node tcp://localhost:26657
    done

