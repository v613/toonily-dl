#!/bin/bash

echo -n "Please insert an Toonly URL:"
read URL

toon=$(curl -s $URL)

Title_Line=$(echo "$toon" | grep -E "^<title>")
start_idx=$(expr length "<title>Read ")
end_idx=$(expr length " Manga - Toonily</title>")
title=${Title_Line:$start_idx:-$end_idx}

mkdir "$title"
cd "$title"

chapters=$(echo "$toon" | grep -A1 -w "class=\"wp-manga-chapter" | grep href | cut -d "\"" -f2)
for chapter in $chapters;
do 
	echo "Downloading $chapter"
	chapter_dir=$(echo $chapter | cut -d "/" -f6)
	
	echo "Make directory $chapter_dir"
	mkdir $chapter_dir
	cd $chapter_dir
	
	imgs=$(curl -s $chapter| grep -A1 -E "image-[[:digit:]]" | grep cdn | awk '{print substr($1,1,length($1)-1)}')
	for img in $imgs;
	do
		echo "Saving $img"
		curl -s $img -H 'authority: cdn.toonily.com' -H 'referer: https://toonily.com/' -O

	done

	cd ../
done
