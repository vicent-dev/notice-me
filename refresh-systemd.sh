#!/bin/bash

cp notice-me.service /lib/systemd/system/notice-me.service && systemctl restart notice-me
