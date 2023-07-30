package service

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/akatsukisun2020/go_components/logger"
	pb "github.com/akatsukisun2020/proto_proj/name_hunter"
	"google.golang.org/grpc/grpclog"
)

var gLogger = grpclog.Component("yewu")

func RandomName(ctx context.Context, req *pb.RandomNameReq) (*pb.RandomNameRsp, error) {
	rsp := new(pb.RandomNameRsp)
	logger.DebugContextf(ctx, "req:%v", req)
	selNames := NameByAncient(ctx, req.GetBook(), int(req.GetCount()))
	for _, v := range selNames {
		item := &pb.SelName{
			Name:     v.Name,
			Title:    v.Title,
			Author:   v.Author,
			Book:     v.Book,
			Dynasty:  v.Dynasty,
			Sentence: v.Sentence,
		}
		rsp.Names = append(rsp.Names, item)
	}

	return rsp, nil
}

// SelName 选择的名字
type SelName struct {
	Name     string `json:"name"`     // 名字
	Title    string `json:"title"`    // 标题
	Author   string `json:"author"`   // 作者
	Book     string `json:"book"`     // 文章名
	Dynasty  string `json:"dynasty"`  //朝代
	Sentence string `json:"sentence"` // 名字
}

// NameByAncient 根据古文进行取名
func NameByAncient(ctx context.Context, ancientType string, number int) []*SelName {
	var selNames []*SelName
	for i := 0; i < number*2; i++ { // 尝试生成数量的两倍
		if len(selNames) >= number {
			break
		}
		if selName := genOneNameByAncient(ancientType); selName != nil {
			selNames = append(selNames, selName)
		}
	}

	jsondata, _ := json.Marshal(selNames)
	logger.InfoContextf(ctx, "in NameByAncient, ancientType:%s, number:%d, selNames:\n%s\n", ancientType, number, string(jsondata))

	return selNames
}

type article struct {
	Content string `json:"content"` // 内容
	Title   string `json:"title"`   // 标题
	Author  string `json:"author"`  // 作者
	Book    string `json:"book"`    // 文章名
	Dynasty string `json:"dynasty"` //朝代
}

type book struct {
	Articles []*article
}

type ancientLoader struct {
	ancientBooks map[string]*book // 不同的书名
}

func GetBookList() []string {
	return []string{"chuci", "cifu", "gushi", "shijing", "songci", "yuefu", "tangshi"} // 支持的书名
}

var gAncientLoader *ancientLoader

// InitAncientLoader 初始化
func InitAncientLoader() {
	ancientLoader := &ancientLoader{
		ancientBooks: make(map[string]*book),
	}

	books := GetBookList()
	for _, b := range books {
		// 读取json文件
		content, err := os.ReadFile("data/" + b + ".json")
		if err != nil {
			logger.DebugContextf(context.Background(), "ReadFile error, err:%v\n", err)
			continue
		}
		var articles []*article
		err = json.Unmarshal(content, &articles)
		if err != nil {
			logger.DebugContextf(context.Background(), "Unmarshal error, err:%v\n", err)
			continue
		}

		ancientLoader.ancientBooks[b] = &book{
			Articles: articles,
		}
	}

	gAncientLoader = ancientLoader
}

// genOneNameByAncient 获得一个名字
func genOneNameByAncient(ancientType string) *SelName {
	if gAncientLoader == nil || gAncientLoader.ancientBooks[ancientType] == nil {
		return nil
	}

	ancientBook := gAncientLoader.ancientBooks[ancientType]
	// (1) 随机获取一个文章
	article := ancientBook.Articles[GenerateRandnum(len(ancientBook.Articles))]

	// (2) 文章切分,随机获取一个句子
	sentences := splitSentence(article.Content)
	if len(sentences) == 0 {
		return nil
	}
	originSentence := sentences[GenerateRandnum(len(sentences))]
	sentence := cleanBadChar(originSentence)
	if len(sentence) < 2 {
		return nil
	}

	// (3) 随机抽取句子中的两个字作为名字
	name := getTwoChar(strings.Split(sentence, ""))
	return &SelName{
		Name:     name,
		Title:    article.Title,
		Author:   article.Author,
		Book:     article.Book,
		Dynasty:  article.Dynasty,
		Sentence: originSentence,
	}
}

// GenerateRandnum 产生 [0:max)之间的随机数
func GenerateRandnum(max int) int {
	rand.Seed(time.Now().UnixNano())
	randNum := rand.Intn(max)
	return randNum
}

func formatStr(str string) string {
	re := regexp.MustCompile(`(\s|　|”|“){1,}|<br>|<p>|<\/p>|\(.+\)`)
	res := re.ReplaceAllString(str, "")
	return res
}

func splitSentence(content string) []string {
	if content == "" {
		return []string{}
	}
	str := formatStr(content)
	str = strings.ReplaceAll(str, "！", "|")
	str = strings.ReplaceAll(str, "。", "|")
	str = strings.ReplaceAll(str, "？", "|")
	str = strings.ReplaceAll(str, "；", "|")
	str = strings.TrimSuffix(str, "|")
	arr := strings.Split(str, "|")
	res := []string{}
	for _, item := range arr {
		if len(item) >= 2 {
			res = append(res, item)
		}
	}
	return res
}

func cleanBadChar(str string) string {
	badChars := strings.Split("，胸鬼懒禽鸟鸡我邪罪凶丑仇鼠蟋蟀淫秽妹狐鸡鸭蝇悔鱼肉苦犬吠窥血丧饥女搔父母昏狗蟊疾病痛死潦哀痒害蛇牲妇狸鹅穴畜烂兽靡爪氓劫鬣螽毛婚姻匪婆羞辱虐乱", "")
	// fmt.Printf("badChars:%v\n", badChars)
	res := ""
	for _, char := range str {
		if !contains(badChars, string(char)) {
			res += string(char)
		}
	}
	return res
}

func contains(arr []string, s string) bool {
	for _, item := range arr {
		if item == s {
			return true
		}
	}
	return false
}

func getTwoChar(arr []string) string {
	len := len(arr)
	first := GenerateRandnum(len)
	second := GenerateRandnum(len)
	cnt := 0
	for second == first {
		second = rand.Intn(len)
		cnt++
		if cnt > 100 {
			break
		}
	}
	if first <= second {
		return arr[first] + arr[second]
	} else {
		return arr[second] + arr[first]
	}
}
