for i in `seq 1 1000`
do
  ./goken-grep2 mattn CONTRIBUTORS 1>/dev/null 2>/dev/null
done
