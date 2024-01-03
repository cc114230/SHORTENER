package logic

import (
	"SHORTENER/internal/common/errorx"
	"SHORTENER/internal/svc"
	"SHORTENER/internal/types"
	"SHORTENER/model"
	"SHORTENER/pkg/base62"
	"SHORTENER/pkg/connect"
	"SHORTENER/pkg/md5"
	"SHORTENER/pkg/urltool"
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ConvertLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConvertLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConvertLogic {
	return &ConvertLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Convert 转链：输入一个长链接-->转为短链接
func (l *ConvertLogic) Convert(req *types.ConvertRequset) (resp *types.ConvertResponse, err error) {
	// 1.校验输入的数据
	// 1.1数据不能为空
	//if len(req.LongUrl) == 0 {}
	// 使用validator包来做参数校验 在handler层那边
	// 1.2输入的长链接是能请求通的网址
	//http.Get(req.LongUrl)
	if ok := connect.Get(req.LongUrl); !ok {
		//return nil, errors.New("无效的链接")
		return nil, errorx.NewCodeError(errorx.InvalidUrl, "无效的链接", nil)
	}
	// 1.3判断之前是否已经转链过（数据库中是否已存在该长链接）
	// 1.3.1 给长链接生成md5
	md5Value := md5.Sum([]byte(req.LongUrl))
	// 1.3.2 拿md5值去数据库查是否存在
	u, err := l.svcCtx.ShortUrlModel.FindOneByMd5(l.ctx, sql.NullString{String: md5Value, Valid: true})
	if err != sqlx.ErrNotFound {
		if err == nil {
			//return nil, fmt.Errorf("该链接已被转为%s", u.Surl.String)
			return nil, errorx.NewCodeError(errorx.IsAlreadyConvert, "该链接已被转链", l.svcCtx.Config.ShortDomain+"/"+u.Surl.String)
		}
		logx.Errorw("ShortUrlModel.FindOneByMd5 failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}

	// 1.4输入的不能是一个短链接（避免循环转链）
	// 输入的是一个完整的url xxx.cn/1jxa2?name=cl
	basePath, err := urltool.GetBasePath(req.LongUrl)
	if err != nil {
		logx.Errorw("urltool.Parse failed", logx.LogField{
			Key: "lurl", Value: req.LongUrl}, logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	_, err = l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{String: basePath, Valid: true})
	if err != sqlx.ErrNotFound {
		if err == nil {
			//return nil, errors.New("该链接已经是短链了")
			return nil, errorx.NewCodeError(errorx.IsAlreadyShortUrl, "该链接已经是短链了", nil)
		}
		logx.Errorw("ShortUrlModel.FindOneBySurl failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	var short string
	for {
		// 2.取号
		seq, err := l.svcCtx.Sequence.Next()
		if err != nil {
			logx.Errorw("sequence.Next failed", logx.LogField{Key: "err", Value: err.Error()})
			return nil, err
		}
		fmt.Println(seq)
		// 3.号码转短链
		// 3.1 安全性 方法：打乱62位字符
		// 3.2 避免特殊词如health api 脏话等等
		short = base62.Int2String(seq)
		if _, ok := l.svcCtx.ShortUrlBlackList[short]; !ok {
			break // 生成不在黑名单里的短链接就跳出for
		}
	}
	// 4.存储长短链接的映射关系
	if _, err := l.svcCtx.ShortUrlModel.Insert(l.ctx, &model.ShortUrlMap{
		Lurl: sql.NullString{String: req.LongUrl, Valid: true},
		Md5:  sql.NullString{String: md5Value, Valid: true},
		Surl: sql.NullString{String: short, Valid: true},
	}); err != nil {
		logx.Errorw("ShortUrlModel.Insert failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}
	// 4.2 将生成的短链接加到布隆过滤器中
	if err := l.svcCtx.Filter.Add([]byte(short)); err != nil {
		logx.Errorw("Filter.Add failed", logx.LogField{Key: "err", Value: err.Error()})
		return nil, err
	}

	// 5.返回响应
	// 5.1 返回的是短域名+短链接
	shortUrl := l.svcCtx.Config.ShortDomain + "/" + short
	return &types.ConvertResponse{
		ShortUrl: shortUrl,
	}, nil
}
