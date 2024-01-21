#!/bin/bash

echo -n "Please insert an Toonly URL:"
read -r URL

toon=$(curl -s "$URL")

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
	
	echo "Downloading $chapter"
	imgs=$(curl -s "$chapter"| grep -A1 -E "image-[[:digit:]]" | grep cdn | awk '{print substr($1,1,length($1)-1)}')
	for img in $imgs;
	do
		echo "Saving $img"
		wget --quiet --header 'authority: cdn.toonily.com' --header 'referer: https://toonily.com/' --continue "$img"
	done

	# Exit from chapter directory
	cd ../
done
