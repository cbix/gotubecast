#This is also a configo file.
#SSH credentials can be modified here too
# SSH needs also  putty on windows ::))))))))))
#Defaultraspbian ssh credentials with the MY rasp's ip on MY router (enable ssh if you don't already have and check it's ip):
set login=pi@192.168.1.79
set pass=raspberry
set app_path=youtube/gotubecast/examples/
set app_name=vlcCostum.sh
ECHO \n
ECHO IMMA GET SOME YOUTUBE ACTION ON U LIL PI
putty.exe -ssh %login% -pw %pass% -m .\configpipath.txt