#!/bin/bash
set -e

total_args=$#
last_arg=${!total_args}

if [[ $total_args -eq 0 || -z "$last_arg" ]]; then
	echo "invalid argument: URL"
	echo "example: toonily-dl.sh https://toonily.com/webtoon/amazing-manga/" 
	exit 1
fi
toon=$(curl -s "$last_arg")

Title_Line=$(echo "$toon" | grep -E "^<title>")
start_idx=12 # 12 <== `<title>Read `
end_idx=24   # 24 <== ` Manga - Toonily</title>`
title=${Title_Line:$start_idx:-$end_idx}
title=$(echo "$title"|sed 's/&#8217;/’/g'|sed 's/&#8230;/.../g')

if [ ! -d "$title" ]; then
	mkdir "$title"
fi
cd "$title" || exit
echo "Download: $title"

if [ ! -e "cover.jpg" ]; then
	wget -qc "$(echo "$toon" | grep -oP 'data-src="\K[^"]+')" -O "cover.jpg"
fi

while getopts ":dc:" flag; do
	case ${flag} in
		c) flag_chapter=$OPTARG	;;
		d) description=$(echo "$toon"|grep -A1 -w "summary__content"|grep "<p>")
		   description=$(echo "$description"|sed -n 's|<p>\(.*\)</p>|\1|p')
		   description=$(echo "$description"|sed 's/&#8217;/’/g'|sed 's/&#8230;/.../g')
		   echo "$description" > "description.txt"
		   ;; 
		*)  echo "Usage: toonily-dl.sh [options] <URL>"; exit 1	;;
	esac
done

IFS=':' read -r -a chapterRange <<< "$flag_chapter"
if [ -z "${chapterRange[0]}" ]; then chapterRange[0]="0"; fi
if [ -z "${chapterRange[1]}" ]; then chapterRange[1]="1000000"; fi

chapters=$(echo "$toon" | grep -A1 -w "class=\"wp-manga-chapter" | grep href | cut -d "\"" -f2)
for chapter in $chapters;
do
	chapter_dir=$(echo "$chapter" | cut -d "/" -f6)
	idx=$(echo "$chapter_dir"|cut -d "-" -f2)

	if [ ! "${idx//[0-9]}" = "" ]; then
		echo "Unexpected chapter '$idx' is about to be downloaded"
	elif [ ! "$idx" -ge "${chapterRange[0]}" ] && [ ! "$idx" -le "${chapterRange[1]}" ]; then
		continue
	elif [ ! "$idx" = "${chapterRange[0]}" ] && [ ! "${chapterRange[1]}" = 0 ]; then
		continue
	fi

	if [ "${idx//[0-9]}" = "" ]; then
		chapter_dir=$(printf "chapter-%03d" "$idx")
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
	echo "Downloaded $(find . -type f|wc -l) file(s)"

	# Exit from chapter directory
	cd ../
done
