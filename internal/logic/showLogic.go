package logic

import (
	"SHORTENER/internal/common/errorx"
	"SHORTENER/internal/svc"
	"SHORTENER/internal/types"
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShowLogic {
	return &ShowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShowLogic) Show(req *types.ShowRequset) (resp *types.ShowResponse, err error) {
	// 查看短链接，qimi/kifjiaef -> 重定向到真实的链接
	// 1. 根据短链接查询到原始的长链接
	// 1.0 使用布隆过滤器，不存在的链接直接返回404，不需要往后处理
	// a.基于内存版本，服务重启之后就没了,所以每次重启都要重新加载已有的短链接
	// b.基于redis版本，go-zero自带
	exist, err := l.svcCtx.Filter.Exists([]byte(req.ShortUrl))
	if err != nil {
		logx.Errorw("Filter.Exists failed", logx.LogField{Value: err.Error(), Key: "err"})
		return nil, err
	}
	// 不存在短链接直接返回
	if !exist {
		//return nil, errors.New("404")
		return nil, errorx.NewCodeError(errorx.PageNotFound, "404", nil)
	}
	fmt.Println("开始查询缓存和DB...")
	// 1.1 查询数据库之前可增加缓存层
	// go-zero自带single flight合并请求，解决缓存击穿问题
	u, err := l.svcCtx.ShortUrlModel.FindOneBySurl(l.ctx, sql.NullString{String: req.ShortUrl, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			//return nil, errors.New("404")
			return nil, errorx.NewCodeError(errorx.PageNotFound, "404", nil)
		}
		logx.Errorw("ShortUrlModel.FindOneBySurl failed", logx.LogField{Value: err.Error(), Key: "err"})
		return nil, err
	}

	// 2. 返回重定向响应 放到了handler层处理
	return &types.ShowResponse{
		LongUrl: u.Lurl.String,
	}, nil

}
