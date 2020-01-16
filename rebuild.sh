#!/bin/sh
# build.sh
# This script for build go projects.
# @version    170227:1
#

#[get path]

current_path=$(cd "$(dirname "$0")"; pwd)

cd "${current_path}/../"
parent_path=$(cd "$(dirname "$0")"; pwd)

project_name=$(basename "$parent_path")
echo "-[ rebuild project name: $project_name ]-"
#rm es_api ../bin/es_api
#go build -o es_api
#cp es_api ../bin

build_target_file=${current_path}/${project_name}
bin_file=${parent_path}/bin/${project_name}

echo "-[ clear $project_name project file ]-"

if test -e $build_target_file
then
    rm $build_target_file
    echo "-[ rm file $build_target_file ]-"
fi

if test -e $bin_file
then
    rm $bin_file
    echo "-[ rm file $bin_file ]-"
fi

echo "-[ project : $project_name : build start ]-"
cd ${current_path}
go build -o $project_name

if [ ! -d ${parent_path}/bin/ ]; then mkdir ${parent_path}/bin/; fi

cp ${build_target_file} ${parent_path}/bin/

#echo $project_name
#echo $build_target_file
#echo $bin_file

echo "-[restart service]-"

if [ -d $parent_path/config/systemctl ]; then 
    cd $parent_path/config/systemctl
else 
   echo "-[no $parent_path/conf/systemctl ]-"
fi

if [ -e $parent_path/config/systemctl/$project_name.service ]; then 
   systemctl stop $project_name.service
   systemctl start $project_name.service
   echo "-[restart $project_name service success]-"
else 
   echo "-[no  $parent_path/conf/systemctl/$project_name.service ]-"
fi

echo ""
echo ""
echo "-[netstat -antp|grep $project_name ]-"
result=`netstat -antp|grep $project_name`
echo $result

echo ""
echo ""

echo "done."