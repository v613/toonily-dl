 #!/bin/bash

if [[ $# -eq 0 || -z "$1" ]]; then
	echo "invalid argument: URL"
	echo "example: toonily-dl.sh https://toonily.com/webtoon/amazing-manga/" 
	exit 1
fi
toon=$(curl -s "$1")

Title_Line=$(echo "$toon" | grep -E "^<title>")
start_idx=$(expr length "<title>Read ")
end_idx=$(expr length " Manga - Toonily</title>")
title=${Title_Line:$start_idx:-$end_idx}

if [ ! -d "$title" ]; then
	mkdir "$title"
fi
cd "$title" || exit
echo "Download: $title"

if [ ! -e "cover.jpg" ]; then
	wget -qc $(echo "$toon" | grep -oP 'data-src="\K[^"]+') -o "cover.jpg"
fi

while getopts "c:" flag; do
	case ${flag} in
		-c)  flag_chapter=$OPTARG	;;
		*)  echo "Usage: toonily-dl.sh [options] <URL>"; exit 1	;;
	esac
done

# while [[ "$#" -gt 0 ]]; do
#     case $1 in
#         -c)
#             if [[ $2 == *":"* ]]; then
#                 start=$(echo $2 | cut -d':' -f1)
#                 end=$(echo $2 | cut -d':' -f2)
#                 echo "Range detected: $start to $end"
#             elif [[ $2 == *":" ]]; then
#                 start=""
#                 end=$(echo $2 | cut -d':' -f1)
#                 echo "End range detected: up to $end"
#             else
#                 echo "Single value detected: $2"
#             fi
#             shift
#             ;;
#         *)
#             echo "Unknown parameter passed: $1"
#             exit 1
#             ;;
#     esac
#     shift
# done

IFS=':' read -r -a chapterRange <<< "$flag_chapter"
if [ -z ${chapterRange[0]} ]; then chapterRange[0]="0"; fi
if [ -z ${chapterRange[1]} ]; then chapterRange[1]="1000000"; fi

chapters=$(echo "$toon" | grep -A1 -w "class=\"wp-manga-chapter" | grep href | cut -d "\"" -f2)
for chapter in $chapters;
do
	chapter_dir=$(echo "$chapter" | cut -d "/" -f6)
	idx=$(echo "$chapter_dir"|cut -d "-" -f2)
	if [ ! $idx -ge ${chapterRange[0]} && ! $idx -le ${chapterRange[1]} ]; then
		continue
	elif [ ! $idx = ${chapterRange[0]} && ! ${chapterRange[1]} = 0 ]; then
		continue
	fi

	if [ ! -d "$chapter_dir" ]; then
		mkdir "$chapter_dir"
	fi
	cd "$chapter_dir" || exit
	
	echo "Working on $chapter_dir"
	imgs=$(curl -s "$chapter"| grep -A1 -E "image-[[:digit:]]" | grep cdn | awk '{print substr($1,1,length($1)-1)}')
	for img in $imgs;
	do
		wget --quiet --header 'authority: cdn.toonily.com' --header 'referer: https://toonily.com/' --continue "$img"
	done
	echo "Downloaded $(ls -1|wc -l) file(s)"

	# Exit from chapter directory
	cd ../
done
