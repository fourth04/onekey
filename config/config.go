package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var configFile string
var Cfg map[string]interface{}

func initFlag() {
	flag.StringVar(&configFile, "c", "", "config file.")
	// 修改提示信息
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "\nUsage: %s <config>\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NFlag() < 1 {
		flag.Usage()
		log.Fatalln("输入参数错误，请参考正确参数格式!")
	}
}

func initConfig() {
	log.Println("正在加载配置文件...")
	file, err := os.OpenFile(configFile, os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalln("配置文件读取失败，请检查配置文件存放位置及读取权限!\n", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	err = dec.Decode(&Cfg)
	if err != nil {
		log.Fatalln("配置文件解析失败, 请检查配置文件格式!\n", err)
	}
	log.Println("加载配置文件成功！")
}

func init() {
	initFlag()
	initConfig()
}
