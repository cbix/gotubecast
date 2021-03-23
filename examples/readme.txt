"Documentation"!

Files and what they're for:

On the raspberry:
  -dumb.sh      - default implementation in shell with gotubecast;
  -raspi.sh     - implementation in shell with gotubecast, omxplayer and youtube-dl;
  -vlcCostum.sh - implementation in shell with gotubecast and cvlc (vlc without GUI);
  
Now if you want to automate the process, with Windows:
  -youtubePI.bat           - auto open ssh and run commands on configyoutubepipath.txt
  -configyoutubepipath.txt - cd into thhe right folders and run 
  Needs:
      - Enable ssh on your raspberry and discover it's ip.
      - Edit youtubePI.bat with the ip for ssh and it's credentials if you have changed them already.
      - Edit configyoutubepipath.txt with the commands on ssh that you might need to open vlcCostum.sh / raspi.sh if you want.
      
      
-BearkillerPT
  
