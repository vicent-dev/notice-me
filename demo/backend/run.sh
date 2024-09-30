i=0

while [ $i -lt 6 ]; do # 6 ten-second intervals in 1 minute
  ./notice-me-demo & #run your command
  sleep 10
  i=$(( i + 1 ))
done

