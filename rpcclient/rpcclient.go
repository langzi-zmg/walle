package rpcclient

import (
	"context"
	"fmt"
	"gitlab.wallstcn.com/wscnbackend/ivankaprotocol/xinge"
	"time"
	"github.com/micro/go-micro"
	ivksvc "gitlab.wallstcn.com/wscnbackend/ivankastd/service"
	"cradle/walle/common"

	"cradle/walle/client"
	"strings"
	"math/rand"
)
var push xinge.PushApiClient

func Init(svc micro.Service) {
	push = xinge.NewPushApiClient("gitlab.wallstcn.com.xinge", svc.Client())
}

func StartService() {
	svc := ivksvc.NewService(common.GlobalConf.Micro)
	svc.Init()
	Init(svc)
}
func ClientSendEmail() {

	client.GetQyUsers()
	client.GetGitlabUsers()
	client.FileSaveRedis()


	leaveUser := make(map[string]common.GitlabUser)
	var leaveUserList []string

	//QyUsers and GitlabUsers
	for key, val := range client.GitlabEmailMap {
		if _, ok := client.QyEmailMap[key]; !ok && client.GitlabEmailMap[key].External == false && strings.Contains(key, "wallstreetcn.com") && client.GitlabEmailMap[key].State == "active" && !strings.Contains(client.GitlabEmailMap[key].Name, "junzhi") && client.GitlabEmailMap[key].Name != "wallstreetcn" {
			fmt.Println("Users who need to be blocked on gitlab: " + val.Name + "  " + val.Email)
			//println(client.BlockGitlabUsers(val.Id))
			leaveUser[key] = val
			leaveUserList = append(leaveUserList, key)

		}
	}

	//QyUsers and sshkey
	redisList := client.GetRedisClient().Keys("*wall*").Val()
	var leaverUserPublish []string
	for _, val := range redisList {
		if _, ok := client.QyEmailMap[val[1:]]; !ok {
			fmt.Println("Users who need to be deleted on publish machine: " + val[1:])
			leaverUserPublish = append(leaverUserPublish, val[1:])
		}
	}

	//ssl info
	var publicVar client.PublicVar
	var sslVar client.SslVar
	publicVar.Action = "CertGetList"
	publicVar.SecretId = client.SecretId
	publicVar.SignatureMethod = "HmacSHA256"
	publicVar.Nonce = fmt.Sprintf("%d", func() int {
		rand.Seed(time.Now().Unix())
		randNum := rand.Intn(10000000)
		return randNum
	}())
	publicVar.Timestamp = fmt.Sprintf("%d", time.Now().Unix())
	publicVar.Region = "ab-shanghai"
	sslVar.Page = "1"
	sslInfo,_ := client.GetSslInfo(publicVar, sslVar)

	//user info
	var content string
	for _, val := range leaveUserList {
		content = content + "Users who need to be deleted on gitlab: " + val +"\n"
	}
	for _, val := range leaverUserPublish {
		content = content +"Users who need to be deleted on publish machine: " + val +"\n"
	}

	//grpc server and send mail
	emailParams := new(xinge.EmailParms)
	emailList := []string{"zhangmengge@wallstreetcn.com"}
	emailParams.Titile = "Users who need to be deleted"
	emailParams.Receivers = emailList
	emailParams.Content= content + sslInfo
	//fmt.Println(emailParams.Content)
	fmt.Println("=======================")
	fmt.Println(emailParams.Content)
	fmt.Println(emailParams.Receivers)
	fmt.Println(emailParams.Titile)
	fmt.Println("=========================")
	ctx, _ := context.WithTimeout(context.Background(), (10 * time.Second))
	rsp, err := push.SendEmail(ctx, emailParams)
	if err != nil {
		fmt.Println("email-sending err: ", err.Error())
	}
	fmt.Println(rsp.Status)

}
