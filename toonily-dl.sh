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

chapters=$(echo "$toon" | grep -A1 -w "class=\"wp-manga-chapter" | grep href | cut -d "\"" -f2)
for chapter in $chapters;
do 
	chapter_dir=$(echo "$chapter" | cut -d "/" -f6)
	if [ ! -d "$chapter_dir" ]; then
		echo "Make directory $chapter_dir"
		mkdir "$chapter_dir"
	fi
	cd "$chapter_dir" || exit
	
	echo "Downloading $chapter_dir"
	imgs=$(curl -s "$chapter"| grep -A1 -E "image-[[:digit:]]" | grep cdn | awk '{print substr($1,1,length($1)-1)}')
	for img in $imgs;
	do
		wget --quiet --header 'authority: cdn.toonily.com' --header 'referer: https://toonily.com/' --continue "$img"
	done

	# Exit from chapter directory
	cd ../
done
