package helpers

import (
	"log"
	"crypto/sha256"
	"math/rand"
	"time"
	c "strconv"
	s "strings"
	"fmt"
)

func init(){
	rand.Seed(time.Now().UnixNano())
}

const PAGE_SIZE = 10

const auth_chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
func GenerateAuthToken(length int) string {
	ret := make([]byte, length)
	for i := range ret{
		ret[i] = auth_chars[rand.Int63() % int64(len(auth_chars))]
	}
	return string(ret)
}

func CheckErrorFatal(err error){
	if err != nil {
		log.Fatal(err)
	}
}

func CheckErrorSafe(err error){
	if err != nil {
		log.Fatal(err)
	}
}

func MakePasswordHash (p string) string {
	hasher := sha256.New()
	hasher.Write([]byte(p+"a wonderful bunch of coconuts"))
	return string(hasher.Sum(nil))
}

func CheckPasswordHash(attempt string, hash string) bool{
	att := c.QuoteToASCII(MakePasswordHash(attempt))
	if att == hash {
		return true
	}
	return false
}

func TimeDelta(t time.Time) string {
	delta := time.Since(t)
	sec   := int64(delta.Seconds())
	if sec>86400{//measure in days
		days := int(sec/86400)
		hours:= int((sec%86400)/3600)
		ret  := fmt.Sprintf("%dd, %dh ago", days, hours)
		return ret
	}
	if sec>3600{//measure in hours
		hours:= int((sec%86400)/3600)
		mins := int((sec%3600)/60)
		ret  := fmt.Sprintf("%dh, %dm ago", hours, mins)
		return ret
	}
	if sec>60{//measure in minutes
		mins := int((sec%3600)/60)
		s    := sec%60
		ret  := fmt.Sprintf("%dm, %ds ago", mins, s)
		return ret
	}
	if sec < 0 {
		ret  := "I hate time zones."
		return ret
	}
	ret := fmt.Sprintf("%ds ago", sec)
	return ret
}

func IsFloat(test string) bool {
	_, err := c.ParseFloat(test, 64)
	if err != nil {
		return false
	}
	return true
}

func IsInt(test string) bool {
	_, err := c.ParseInt(test, 10, 64)
	if err != nil {
		return false
	}
	return true
}

var colorNames = [...]string  {"crimson","salmon","red","maroon","orangered","gold","orange","khaki","yellow","lawngreen","limegreeen","lime","green","springgreen","seagreen","olive","cyan","aquamarine","turquoise","teal","lightseagreen","deepskyblue","blue","navy","slateblue","magenta","blueviolet","purple","indigo","hotpink","gray","darkslategray","goldenrod","saddlebrown"}
var colors     = [...]string  {"#dc143c","#fa8072","#ff0000","#800000","#ff4500","#ffd700","#ffa500","#f0e68c","#ffff00","#7cfc00","#32cd32","#00ff00","#008000","#00ff7f","#2e8b57","#808000","#00ffff","#7fffd4","#40e0d0","#008080","#20b2aa","#00bfff","#0000ff","#000080","#6a5acd","#ff00ff","#8a2be2","#800080","#4b0082","#ff69b4","#808080","#2f4f4f","#daa520","#8b4513"}
var icons      = [...]string {"alligator","bear","beaver","bird","bison","bul","camel", "chameleon","chicken","chipmunk","cobra","deer","dolphin","elephant","fish","frog","gazelle","giraffe","gorilla","horse","kangaroo","kiwi","lemur","llama","monkey","moose","mouse","octopus","ostrich","panda","panther","penguin","pyhton","rabbit","ram","thino","squirrel","tiger","turtle","wolf"}

func GenerateUsername() (string, string, string){//returns name, color, icon
	icon := icons[rand.Int63() % int64(len(icons))]
	idx := rand.Int63() % int64(len(colors))
	color := colors[idx]
	cn := colorNames[idx]
	name := s.Title(cn)+s.Title(icon)
	icon += ".png"
	name = c.QuoteToASCII(name)
	name = name[1:len(name)-1]
	return name, color, icon
}


