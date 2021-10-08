//聊天服务器demo

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"strings"

	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/groups"
	"github.com/topfreegames/pitaya/v2/logger"
	"github.com/topfreegames/pitaya/v2/timer"
)

type (
	// Room represents a component that contains a bundle of room related handler
	// like Join/Message
	// Room表示包含一组与Room相关的处理程序的组件
	// /加入/信息
	// room是一個component
	// app 成員是pitaya.Pitaya
	Room struct {
		component.Base
		timer *timer.Timer
		app   pitaya.Pitaya
	}

	// UserMessage represents a message that user sent
	// 表示用户发送的消息
	UserMessage struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	// NewUser message will be received when new user join room
	// 当新用户加入房间时，将收到NewUser消息
	NewUser struct {
		Content string `json:"content"`
	}

	// AllMembers contains all members uid
	// AllMembers包含所有成员uid
	AllMembers struct {
		Members []string `json:"members"`
	}

	// JoinResponse represents the result of joining room
	// JoinResponse表示加入房间的结果
	JoinResponse struct {
		Code   int    `json:"code"`
		Result string `json:"result"`
	}
)

// NewRoom returns a Handler Base implementation
// 返回一个基础的handler的实现
func NewRoom(app pitaya.Pitaya) *Room {
	return &Room{
		app: app,
	}
}

// AfterInit component lifetime callback
// 测试 初始化完成后 组件的生命周期回调
func (r *Room) AfterInit() {
	r.timer = pitaya.NewTimer(time.Minute, func() {
		count, err := r.app.GroupCountMembers(context.Background(), "room")
		logger.Log.Debugf("UserCount: Time=> %s, Count=> %d, Error=> %q", time.Now().String(), count, err)
	})
}

// Join room
// 加入房间
func (r *Room) Join(ctx context.Context, msg []byte) (*JoinResponse, error) {
	//从上下文获取会话
	s := r.app.GetSessionFromCtx(ctx)

	fakeUID := s.ID()                              // just use s.ID as uid !!!
	err := s.Bind(ctx, strconv.Itoa(int(fakeUID))) // binding session uid

	if err != nil {
		return nil, pitaya.Error(err, "RH-000", map[string]string{"failed": "bind"})
	}
	//得到当前所有成员 房间
	uids, err := r.app.GroupMembers(ctx, "room")
	if err != nil {
		return nil, err
	}
	//push给所有成员
	s.Push("onMembers", &AllMembers{Members: uids})
	// notify others
	//通知其他 广播
	r.app.GroupBroadcast(ctx, "chat", "room", "onNewUser", &NewUser{Content: fmt.Sprintf("New user: %s", s.UID())})
	// new user join group
	//广播新用户进入
	r.app.GroupAddMember(ctx, "room", s.UID()) // add session to group

	// on session close, remove it from group
	// 当会话关闭，把会话从组移除
	s.OnClose(func() {
		r.app.GroupRemoveMember(ctx, "room", s.UID())
	})
	//返回加入成功
	return &JoinResponse{Result: "success"}, nil
}

// Message sync last message to all members
// 同步最后一条消息给所有成员
// 广播消息msg
func (r *Room) Message(ctx context.Context, msg *UserMessage) {
	err := r.app.GroupBroadcast(ctx, "chat", "room", "onMessage", msg)
	if err != nil {
		fmt.Println("error broadcasting message", err)
	}
}

var app pitaya.Pitaya

func main() {
	//返回building conf
	conf := configApp()
	//返回builder
	builder := pitaya.NewDefaultBuilder(true, "chat", pitaya.Cluster, map[string]string{}, *conf)
	//添加一个接收器 這裏是一個websocket 監聽3250
	builder.AddAcceptor(acceptor.NewWSAcceptor(":3250"))
	//得到一个组的实例
	builder.Groups = groups.NewMemoryGroupService(*config.NewDefaultMemoryGroupConfig())
	//builder来创建app
	app = builder.Build() //build的工作很多。请仔细查看。包括app主要的 各种服务器创建 。rpc服务创建 工厂类实例化
	//延迟关闭app
	defer app.Shutdown()
	//创建组
	err := app.GroupCreate(context.Background(), "room")

	if err != nil {
		panic(err)
	}

	// rewrite component and handler name
	//重写组件和处理程序名称
	room := NewRoom(app)
	//注册
	app.Register(room,
		component.WithName("room"),
		component.WithNameFunc(strings.ToLower),
	)
	//设置log的标志
	log.SetFlags(log.LstdFlags | log.Llongfile)
	//返回打开聊天网页
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))
	//http监听服务
	go http.ListenAndServe(":3251", nil)
	//启动app
	app.Start()
}

//配置app
func configApp() *config.BuilderConfig {
	conf := config.NewDefaultBuilderConfig()                         //默认的buildconfig
	conf.Pitaya.Buffer.Handler.LocalProcess = 15                     //本地处理。15是什么？
	conf.Pitaya.Heartbeat.Interval = time.Duration(15 * time.Second) //心跳时间间隔
	conf.Pitaya.Buffer.Agent.Messages = 32                           //Messages 不知道什么意思 证书类型
	conf.Pitaya.Handler.Messages.Compression = false                 //压缩消息
	return conf
}
