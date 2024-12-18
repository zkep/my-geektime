package geek

type ProductListRequest struct {
	Desc           bool   `json:"desc" form:"desc"`
	Expire         int    `json:"expire" form:"expire"`
	LastLearn      int    `json:"last_learn" form:"last_learn"`
	LearnStatus    int    `json:"learn_status" form:"learn_status"`
	Prev           int    `json:"prev" form:"prev"`
	Size           int    `json:"size" form:"size"`
	Sort           int    `json:"sort" form:"sort"`
	Type           string `json:"type" form:"type"`
	WithLearnCount int    `json:"with_learn_count" form:"with_learn_count"`
	Page           int    `json:"-" form:"page"`
	PerPage        int    `json:"-"  form:"perPage"`
}

type ProductListResponse struct {
	HasNext bool             `json:"hasNext,omitempty"`
	Count   int              `json:"count,omitempty"`
	Rows    []ProductListRow `json:"rows"`
}

type ProductListRow struct {
	ID            int            `json:"id,omitempty"`
	Title         string         `json:"title,omitempty"`
	Subtitle      string         `json:"subtitle,omitempty"`
	Intro         string         `json:"intro,omitempty"`
	IntroHTML     string         `json:"intro_html,omitempty"`
	Ucode         string         `json:"ucode,omitempty"`
	IsFinish      bool           `json:"is_finish,omitempty"`
	IsVideo       bool           `json:"is_video,omitempty"`
	IsAudio       bool           `json:"is_audio,omitempty"`
	IsDailylesson bool           `json:"is_dailylesson,omitempty"`
	IsUniversity  bool           `json:"is_university,omitempty"`
	IsOpencourse  bool           `json:"is_opencourse,omitempty"`
	IsQconp       bool           `json:"is_qconp,omitempty"`
	IsSale        bool           `json:"is_sale,omitempty"`
	Sale          int            `json:"sale,omitempty"`
	SaleType      int            `json:"sale_type,omitempty"`
	Share         ProductShare   `json:"share,omitempty"`
	Author        ProductAuthor  `json:"author,omitempty"`
	Cover         ProductCover   `json:"cover,omitempty"`
	Article       ProductArticle `json:"article,omitempty"`
}

type ProductShare struct {
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Cover   string `json:"cover,omitempty"`
	Poster  string `json:"poster,omitempty"`
}

type ProductAuthor struct {
	Name      string `json:"name,omitempty"`
	Intro     string `json:"intro,omitempty"`
	Info      string `json:"info,omitempty"`
	Avatar    string `json:"avatar,omitempty"`
	BriefHTML string `json:"brief_html,omitempty"`
	Brief     string `json:"brief,omitempty"`
	AiID      string `json:"ai_id,omitempty"`
}

type ProductCover struct {
	Square            string `json:"square,omitempty"`
	Rectangle         string `json:"rectangle,omitempty"`
	Horizontal        string `json:"horizontal,omitempty"`
	LectureHorizontal string `json:"lecture_horizontal,omitempty"`
	LearnHorizontal   string `json:"learn_horizontal,omitempty"`
	Transparent       string `json:"transparent,omitempty"`
	Color             string `json:"color,omitempty"`
}

type ProductArticle struct {
	ID                int    `json:"id,omitempty"`
	Count             int    `json:"count,omitempty"`
	CountReq          int    `json:"count_req,omitempty"`
	CountPub          int    `json:"count_pub,omitempty"`
	TotalLength       int    `json:"total_length,omitempty"`
	FirstArticleID    int    `json:"first_article_id,omitempty"`
	FirstArticleTitle string `json:"first_article_title,omitempty"`
}

type Extra struct {
	Cost      any `json:"cost,omitempty"`
	RequestID any `json:"request-id,omitempty"`
}

type ProductResponse struct {
	Code  int   `json:"code,omitempty"`
	Error any   `json:"error,omitempty"`
	Extra Extra `json:"extra,omitempty"`
	Data  struct {
		HasExpiringProduct bool `json:"has_expiring_product,omitempty"`
		LearnCount         struct {
			Total int64 `json:"total,omitempty"`
		} `json:"learn_count,omitempty"`
		List []struct {
			Pid             int    `json:"pid,omitempty"`
			Ptype           string `json:"ptype,omitempty"`
			Aid             int    `json:"aid,omitempty"`
			Ctime           int64  `json:"ctime,omitempty"`
			Score           int    `json:"score,omitempty"`
			IsExpire        bool   `json:"is_expire,omitempty"`
			ExpireTime      int    `json:"expire_time,omitempty"`
			LastLearnedTime int    `json:"last_learned_time,omitempty"`
			IsTop           bool   `json:"is_top,omitempty"`
		} `json:"list,omitempty"`
		Articles []any `json:"articles,omitempty"`
		Products []struct {
			ID        int `json:"id,omitempty"`
			Spu       int `json:"spu,omitempty"`
			Ctime     int `json:"ctime,omitempty"`
			Utime     int `json:"utime,omitempty"`
			BeginTime int `json:"begin_time,omitempty"`
			EndTime   int `json:"end_time,omitempty"`
			Price     struct {
				Market    int `json:"market,omitempty"`
				Sale      int `json:"sale,omitempty"`
				SaleType  int `json:"sale_type,omitempty"`
				StartTime int `json:"start_time,omitempty"`
				EndTime   int `json:"end_time,omitempty"`
			} `json:"price,omitempty"`
			IsSale        bool           `json:"is_sale,omitempty"`
			IsGroupbuy    bool           `json:"is_groupbuy,omitempty"`
			IsPromo       bool           `json:"is_promo,omitempty"`
			IsShareget    bool           `json:"is_shareget,omitempty"`
			IsSharesale   bool           `json:"is_sharesale,omitempty"`
			OnlySellInVip bool           `json:"only_sell_in_vip,omitempty"`
			Type          string         `json:"type,omitempty"`
			IsColumn      bool           `json:"is_column,omitempty"`
			IsCore        bool           `json:"is_core,omitempty"`
			IsVideo       bool           `json:"is_video,omitempty"`
			IsAudio       bool           `json:"is_audio,omitempty"`
			IsDailylesson bool           `json:"is_dailylesson,omitempty"`
			IsUniversity  bool           `json:"is_university,omitempty"`
			IsOpencourse  bool           `json:"is_opencourse,omitempty"`
			IsQconp       bool           `json:"is_qconp,omitempty"`
			IsMentor      bool           `json:"is_mentor,omitempty"`
			NavID         int            `json:"nav_id,omitempty"`
			TimeNotSale   int            `json:"time_not_sale,omitempty"`
			Title         string         `json:"title,omitempty"`
			Subtitle      string         `json:"subtitle,omitempty"`
			Intro         string         `json:"intro,omitempty"`
			IntroHTML     string         `json:"intro_html,omitempty"`
			Ucode         string         `json:"ucode,omitempty"`
			IsFinish      bool           `json:"is_finish,omitempty"`
			Share         ProductShare   `json:"share,omitempty"`
			Author        ProductAuthor  `json:"author,omitempty"`
			Cover         ProductCover   `json:"cover,omitempty"`
			Article       ProductArticle `json:"article,omitempty"`
			Seo           struct {
				Keywords []string `json:"keywords,omitempty"`
			} `json:"seo,omitempty"`
			Labels     []int  `json:"labels,omitempty"`
			Unit       string `json:"unit,omitempty"`
			ColumnType int    `json:"column_type,omitempty"`
			Column     struct {
				Unit             string `json:"unit,omitempty"`
				Bgcolor          string `json:"bgcolor,omitempty"`
				UpdateFrequency  string `json:"update_frequency,omitempty"`
				IsPreorder       bool   `json:"is_preorder,omitempty"`
				IsFinish         bool   `json:"is_finish,omitempty"`
				IsIncludePreview bool   `json:"is_include_preview,omitempty"`
				ShowChapter      bool   `json:"show_chapter,omitempty"`
				IsSaleProduct    bool   `json:"is_sale_product,omitempty"`
				StudyService     []int  `json:"study_service,omitempty"`
				Path             struct {
					Desc     string `json:"desc,omitempty"`
					DescHTML string `json:"desc_html,omitempty"`
				} `json:"path,omitempty"`
				IsCamp                   bool   `json:"is_camp,omitempty"`
				CatalogPicURL            string `json:"catalog_pic_url,omitempty"`
				RecommendArticles        any    `json:"recommend_articles,omitempty"`
				RecommendComments        []any  `json:"recommend_comments,omitempty"`
				Ranks                    any    `json:"ranks,omitempty"`
				HotComments              any    `json:"hot_comments,omitempty"`
				HotLines                 any    `json:"hot_lines,omitempty"`
				DisplayType              int    `json:"display_type,omitempty"`
				IntroBgStyle             int    `json:"intro_bg_style,omitempty"`
				CommentTopAds            string `json:"comment_top_ads,omitempty"`
				ArticleFloatQrcodeURL    string `json:"article_float_qrcode_url,omitempty"`
				ArticleFloatAppQrcodeURL string `json:"article_float_app_qrcode_url,omitempty"`
				ArticleFloatQrcodeJump   string `json:"article_float_qrcode_jump,omitempty"`
				InRank                   bool   `json:"in_rank,omitempty"`
			} `json:"column,omitempty"`
			Dl struct {
				Article struct {
					ID            int    `json:"id,omitempty"`
					VideoDuration string `json:"video_duration,omitempty"`
					VideoHot      int    `json:"video_hot,omitempty"`
					CouldPreview  bool   `json:"could_preview,omitempty"`
				} `json:"article,omitempty"`
				TopicIds []any `json:"topic_ids,omitempty"`
			} `json:"dl,omitempty"`
			University struct {
				TotalHour       int    `json:"total_hour,omitempty"`
				Term            int    `json:"term,omitempty"`
				RedirectType    string `json:"redirect_type,omitempty"`
				RedirectParam   string `json:"redirect_param,omitempty"`
				WxQrcode        string `json:"wx_qrcode,omitempty"`
				WxRule          string `json:"wx_rule,omitempty"`
				ServerStartTime int    `json:"server_start_time,omitempty"`
				LecturerHCover  string `json:"lecturer_h_cover,omitempty"`
				Keywords        string `json:"keywords,omitempty"`
			} `json:"university,omitempty"`
			Opencourse struct {
				VideoBg string `json:"video_bg,omitempty"`
				Ad      struct {
					Cover         string `json:"cover,omitempty"`
					CoverWeb      string `json:"cover_web,omitempty"`
					RedirectType  string `json:"redirect_type,omitempty"`
					RedirectParam string `json:"redirect_param,omitempty"`
				} `json:"ad,omitempty"`
				ArticleFav struct {
					Aid     int  `json:"aid,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
					Count   int  `json:"count,omitempty"`
				} `json:"article_fav,omitempty"`
				AuthorHCover string `json:"author_h_cover,omitempty"`
			} `json:"opencourse,omitempty"`
			Qconp struct {
				TopicID      int    `json:"topic_id,omitempty"`
				CoverAppoint string `json:"cover_appoint,omitempty"`
				Article      struct {
					ID            int    `json:"id,omitempty"`
					Cover         string `json:"cover,omitempty"`
					VideoDuration string `json:"video_duration,omitempty"`
					VideoHot      int    `json:"video_hot,omitempty"`
				} `json:"article,omitempty"`
			} `json:"qconp,omitempty"`
			FavQrcode string `json:"fav_qrcode,omitempty"`
			Extra     struct {
				Sub struct {
					Count      int  `json:"count,omitempty"`
					HadDone    bool `json:"had_done,omitempty"`
					CouldOrder bool `json:"could_order,omitempty"`
					AccessMask int  `json:"access_mask,omitempty"`
				} `json:"sub,omitempty"`
				Fav struct {
					Count   int  `json:"count,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
				} `json:"fav,omitempty"`
				Rate struct {
					ArticleCount    int  `json:"article_count,omitempty"`
					ArticleCountReq int  `json:"article_count_req,omitempty"`
					IsFinished      bool `json:"is_finished,omitempty"`
					RatePercent     int  `json:"rate_percent,omitempty"`
					VideoSeconds    int  `json:"video_seconds,omitempty"`
					LastArticleID   int  `json:"last_article_id,omitempty"`
					LastChapterID   int  `json:"last_chapter_id,omitempty"`
					HasLearn        bool `json:"has_learn,omitempty"`
				} `json:"rate,omitempty"`
				Cert struct {
					ID   string `json:"id,omitempty"`
					Type int    `json:"type,omitempty"`
				} `json:"cert,omitempty"`
				Nps struct {
					Min    int    `json:"min,omitempty"`
					Status int    `json:"status,omitempty"`
					URL    string `json:"url,omitempty"`
				} `json:"nps,omitempty"`
				AnyRead struct {
					Total int `json:"total,omitempty"`
					Count int `json:"count,omitempty"`
				} `json:"any_read,omitempty"`
				University struct {
					Status               int    `json:"status,omitempty"`
					ViewStatus           int    `json:"view_status,omitempty"`
					ChargeStatus         int    `json:"charge_status,omitempty"`
					ShareRenewalStatus   int    `json:"share_renewal_status,omitempty"`
					UnlockedStatus       int    `json:"unlocked_status,omitempty"`
					UnlockedChapterIds   []any  `json:"unlocked_chapter_ids,omitempty"`
					UnlockedChapterID    int    `json:"unlocked_chapter_id,omitempty"`
					UnlockedChapterTitle string `json:"unlocked_chapter_title,omitempty"`
					UnlockedArticleCount int    `json:"unlocked_article_count,omitempty"`
					UnlockedNextTime     int    `json:"unlocked_next_time,omitempty"`
					ExpireTime           int    `json:"expire_time,omitempty"`
					IsExpired            bool   `json:"is_expired,omitempty"`
					IsGraduated          bool   `json:"is_graduated,omitempty"`
					HadSub               bool   `json:"had_sub,omitempty"`
					Timeline             []any  `json:"timeline,omitempty"`
					HasWxFriend          bool   `json:"has_wx_friend,omitempty"`
					StartTime            int    `json:"start_time,omitempty"`
					SubTermTitle         string `json:"sub_term_title,omitempty"`
					SubSku               int    `json:"sub_sku,omitempty"`
				} `json:"university,omitempty"`
				Vip struct {
					IsYearCard bool `json:"is_year_card,omitempty"`
					Show       bool `json:"show,omitempty"`
					EndTime    int  `json:"end_time,omitempty"`
				} `json:"vip,omitempty"`
				Appoint struct {
					CouldDo bool `json:"could_do,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
					Count   int  `json:"count,omitempty"`
				} `json:"appoint,omitempty"`
				GroupBuy struct {
					SuccessUcount int    `json:"success_ucount,omitempty"`
					JoinCode      string `json:"join_code,omitempty"`
					CouldGroupbuy bool   `json:"could_groupbuy,omitempty"`
					HadJoin       bool   `json:"had_join,omitempty"`
					Price         int    `json:"price,omitempty"`
					List          []any  `json:"list,omitempty"`
				} `json:"group_buy,omitempty"`
				Sharesale struct {
					OriginalPicColor    string `json:"original_pic_color,omitempty"`
					OriginalPicURL      string `json:"original_pic_url,omitempty"`
					PromoPicColor       string `json:"promo_pic_color,omitempty"`
					PromoPicURL         string `json:"promo_pic_url,omitempty"`
					ShareSalePrice      int    `json:"share_sale_price,omitempty"`
					ShareSaleGuestPrice int    `json:"share_sale_guest_price,omitempty"`
				} `json:"sharesale,omitempty"`
				Promo struct {
					EntTime int `json:"ent_time,omitempty"`
				} `json:"promo,omitempty"`
				Channel struct {
					Is         bool `json:"is,omitempty"`
					BackAmount int  `json:"back_amount,omitempty"`
				} `json:"channel,omitempty"`
				FirstPromo struct {
					Price     int  `json:"price,omitempty"`
					CouldJoin bool `json:"could_join,omitempty"`
				} `json:"first_promo,omitempty"`
				CouponPromo struct {
					CouldJoin bool `json:"could_join,omitempty"`
					Price     int  `json:"price,omitempty"`
				} `json:"coupon_promo,omitempty"`
				Helper []any `json:"helper,omitempty"`
				Tab    struct {
					Comment bool `json:"comment,omitempty"`
					Package bool `json:"package,omitempty"`
				} `json:"tab,omitempty"`
				Modules   []any `json:"modules,omitempty"`
				Cid       int   `json:"cid,omitempty"`
				FirstAids []any `json:"first_aids,omitempty"`
				StudyPlan struct {
					ID              int `json:"id,omitempty"`
					DayNums         int `json:"day_nums,omitempty"`
					ArticleNums     int `json:"article_nums,omitempty"`
					LearnedWeekNums int `json:"learned_week_nums,omitempty"`
					Status          int `json:"status,omitempty"`
				} `json:"study_plan,omitempty"`
				CateID   int    `json:"cate_id,omitempty"`
				CateName string `json:"cate_name,omitempty"`
				GroupTag struct {
					IsRecommend     bool `json:"is_recommend,omitempty"`
					IsRecentlyLearn bool `json:"is_recently_learn,omitempty"`
					IsTop           bool `json:"is_top,omitempty"`
				} `json:"group_tag,omitempty"`
				FirstAward struct {
					Show          bool   `json:"show,omitempty"`
					Talks         int    `json:"talks,omitempty"`
					Reads         int    `json:"reads,omitempty"`
					Amount        int    `json:"amount,omitempty"`
					ExpireTime    int    `json:"expire_time,omitempty"`
					RedirectType  string `json:"redirect_type,omitempty"`
					RedirectParam string `json:"redirect_param,omitempty"`
				} `json:"first_award,omitempty"`
				VipPromo struct {
					DiscountLevel int `json:"discount_level,omitempty"`
					DiscountPrice int `json:"discount_price,omitempty"`
					MinLevel      int `json:"min_level,omitempty"`
					Rules         any `json:"rules,omitempty"`
				} `json:"vip_promo,omitempty"`
				IsTgoTicket bool `json:"is_tgo_ticket,omitempty"`
				BPack       any  `json:"b_pack,omitempty"`
				PSkus       any  `json:"p_skus,omitempty"`
			} `json:"extra,omitempty"`
			AvailableCoupons any    `json:"available_coupons,omitempty"`
			InPvip           int    `json:"in_pvip,omitempty"`
			IsJoinCvip       int    `json:"is_join_cvip,omitempty"`
			ColumnBadge      string `json:"column_badge,omitempty"`
			HideCopyright    bool   `json:"hide_copyright,omitempty"`
		} `json:"products,omitempty"`
		Page struct {
			More   bool `json:"more,omitempty"`
			Total  int  `json:"total,omitempty"`
			Score  int  `json:"score,omitempty"`
			Score0 int  `json:"score0,omitempty"`
		} `json:"page,omitempty"`
	} `json:"data,omitempty"`
}

type DowloadRequest struct {
	Pid int64 `json:"pid,omitempty" binding:"required"`
	Ids any   `json:"ids,omitempty"`
}

type DowloadResponse struct {
	JobID string `json:"job_id,omitempty"`
}

type ArticlesListRequest struct {
	Cid     string `json:"cid" form:"cid" binding:"required"`
	Size    int    `json:"size" form:"size"`
	Prev    int    `json:"prev" form:"prev"`
	Order   string `json:"order" form:"order"`
	Sample  bool   `json:"sample" form:"sample"`
	Page    int    `json:"-" form:"page"`
	PerPage int    `json:"-"  form:"perPage"`
}

type ArticlesResponse struct {
	Error any `json:"error,omitempty"`
	Extra any `json:"extra,omitempty"`
	Data  struct {
		List []struct {
			ID                  int64  `json:"id,omitempty"`
			HadViewed           bool   `json:"had_viewed,omitempty"`
			ArticleTitle        string `json:"article_title,omitempty"`
			ArticleCover        string `json:"article_cover,omitempty"`
			ArticleSubtitle     string `json:"article_subtitle,omitempty"`
			VideoCover          string `json:"video_cover,omitempty"`
			AuthorName          string `json:"author_name,omitempty"`
			AuthorIntro         string `json:"author_intro,omitempty"`
			AudioDownloadURL    string `json:"audio_download_url,omitempty"`
			AudioSize           int    `json:"audio_size,omitempty"`
			AudioTime           string `json:"audio_time,omitempty"`
			ArticleCouldPreview bool   `json:"article_could_preview,omitempty"`
			ChapterID           string `json:"chapter_id,omitempty"`
			ColumnHadSub        bool   `json:"column_had_sub,omitempty"`
			ReadingTime         int    `json:"reading_time,omitempty"`
			IsFinished          bool   `json:"is_finished,omitempty"`
			Subtitles           []any  `json:"subtitles,omitempty"`
			IPAddress           string `json:"ip_address,omitempty"`
			IncludeAudio        bool   `json:"include_audio,omitempty"`
			IsVideoPreview      bool   `json:"is_video_preview,omitempty"`
			ArticleSummary      string `json:"article_summary,omitempty"`
			RatePercent         int    `json:"rate_percent,omitempty"`
			VideoSize           int    `json:"video_size,omitempty"`
			VideoID             string `json:"video_id,omitempty"`
			ColumnSku           int    `json:"column_sku,omitempty"`
			Offline             struct {
				FileName    string `json:"file_name,omitempty"`
				DownloadURL string `json:"download_url,omitempty"`
			} `json:"offline,omitempty"`
			VideoTime  string `json:"video_time,omitempty"`
			IsRequired bool   `json:"is_required,omitempty"`
			Rate       struct {
				Num1 struct {
					CurVersion     int  `json:"cur_version,omitempty"`
					MaxRate        int  `json:"max_rate,omitempty"`
					CurRate        int  `json:"cur_rate,omitempty"`
					IsFinished     bool `json:"is_finished,omitempty"`
					TotalRate      int  `json:"total_rate,omitempty"`
					LearnedSeconds int  `json:"learned_seconds,omitempty"`
				} `json:"1,omitempty"`
				Num2 struct {
					CurVersion     int  `json:"cur_version,omitempty"`
					MaxRate        int  `json:"max_rate,omitempty"`
					CurRate        int  `json:"cur_rate,omitempty"`
					IsFinished     bool `json:"is_finished,omitempty"`
					TotalRate      int  `json:"total_rate,omitempty"`
					LearnedSeconds int  `json:"learned_seconds,omitempty"`
				} `json:"2,omitempty"`
				Num3 struct {
					CurVersion     int  `json:"cur_version,omitempty"`
					MaxRate        int  `json:"max_rate,omitempty"`
					CurRate        int  `json:"cur_rate,omitempty"`
					IsFinished     bool `json:"is_finished,omitempty"`
					TotalRate      int  `json:"total_rate,omitempty"`
					LearnedSeconds int  `json:"learned_seconds,omitempty"`
				} `json:"3,omitempty"`
			} `json:"rate,omitempty"`
			Score        int64 `json:"score,omitempty"`
			ArticleCtime int   `json:"article_ctime,omitempty"`
		} `json:"list,omitempty"`
		Page struct {
			Count int64 `json:"count,omitempty"`
			More  bool  `json:"more,omitempty"`
		} `json:"page,omitempty"`
	} `json:"data,omitempty"`
	Code int `json:"code,omitempty"`
}

type ArticlesListResponse struct {
	Count int64             `json:"count"`
	Rows  []ArticlesListRow `json:"rows"`
}

type ArticlesListRow struct {
	ID               int64  `json:"id,omitempty"`
	ArticleTitle     string `json:"article_title,omitempty"`
	ArticleCover     string `json:"article_cover,omitempty"`
	ArticleSubtitle  string `json:"article_subtitle,omitempty"`
	VideoCover       string `json:"video_cover,omitempty"`
	VideoTime        string `json:"video_time,omitempty"`
	VideoSize        int    `json:"video_size,omitempty"`
	AudioSize        int    `json:"audio_size,omitempty"`
	AudioTime        string `json:"audio_time,omitempty"`
	AudioDownloadURL string `json:"audio_download_url,omitempty"`
	ArticleSummary   string `json:"article_summary,omitempty"`
	AuthorName       string `json:"author_name,omitempty"`
	AuthorIntro      string `json:"author_intro,omitempty"`
	ArticleCtime     int    `json:"article_ctime,omitempty"`
}

type ArticlesInfoRequest struct {
	Id int64 `json:"id,omitempty" form:"id"`
}

type ArticleInfoResponse struct {
	Code  int         `json:"code,omitempty"`
	Data  ArticleData `json:"data,omitempty"`
	Error any         `json:"error,omitempty"`
	Extra Extra       `json:"extra,omitempty"`
}

type ArticleData struct {
	Info    ArticleInfo `json:"info,omitempty"`
	Product struct {
		ID         int    `json:"id,omitempty"`
		Title      string `json:"title,omitempty"`
		University struct {
			RedirectType  string `json:"redirect_type,omitempty"`
			RedirectParam string `json:"redirect_param,omitempty"`
		} `json:"university,omitempty"`
		Extra struct {
			Sub struct {
				HadDone    bool `json:"had_done,omitempty"`
				AccessMask int  `json:"access_mask,omitempty"`
			} `json:"sub,omitempty"`
		} `json:"extra,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"product,omitempty"`
	FreeGet    bool `json:"free_get,omitempty"`
	IsFullText bool `json:"is_full_text,omitempty"`
}

type ArticleInfo struct {
	Id           int    `json:"id,omitempty"`
	Pid          int    `json:"pid,omitempty"`
	Type         int    `json:"type,omitempty"`
	ChapterID    int    `json:"chapter_id,omitempty"`
	ChapterTitle string `json:"chapter_title,omitempty"`
	Title        string `json:"title,omitempty"`
	Subtitle     string `json:"subtitle,omitempty"`
	ShareTitle   string `json:"share_title,omitempty"`
	Summary      string `json:"summary,omitempty"`
	Ctime        int    `json:"ctime,omitempty"`
	Cover        struct {
		Default string `json:"default,omitempty"`
	} `json:"cover,omitempty"`
	Author struct {
		Name   string `json:"name,omitempty"`
		Avatar string `json:"avatar,omitempty"`
	} `json:"author,omitempty"`
	Audio struct {
		Title       string   `json:"title,omitempty"`
		Dubber      string   `json:"dubber,omitempty"`
		DownloadURL string   `json:"download_url,omitempty"`
		Md5         string   `json:"md5,omitempty"`
		Size        int      `json:"size,omitempty"`
		Time        string   `json:"time,omitempty"`
		TimeArr     []string `json:"time_arr,omitempty"`
		URL         string   `json:"url,omitempty"`
	} `json:"audio,omitempty"`
	Video struct {
		ID       string `json:"id,omitempty"`
		Duration int    `json:"duration,omitempty"`
		Cover    string `json:"cover,omitempty"`
		Width    int    `json:"width,omitempty"`
		Height   int    `json:"height,omitempty"`
		Size     int    `json:"size,omitempty"`
		Time     string `json:"time,omitempty"`
		Medias   []struct {
			Size    int    `json:"size,omitempty"`
			Quality string `json:"quality,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"medias,omitempty"`
		HlsVid    string `json:"hls_vid,omitempty"`
		HlsMedias []struct {
			Size    int    `json:"size,omitempty"`
			Quality string `json:"quality,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"hls_medias,omitempty"`
		Subtitles []any `json:"subtitles,omitempty"`
		Tips      []any `json:"tips,omitempty"`
	} `json:"video,omitempty"`
	VideoPreview struct {
		Duration int `json:"duration,omitempty"`
		Medias   []struct {
			Size    int    `json:"size,omitempty"`
			Quality string `json:"quality,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"medias,omitempty"`
	} `json:"video_preview,omitempty"`
	VideoPreviews []struct {
		Duration int `json:"duration,omitempty"`
		Medias   []struct {
			Size    int    `json:"size,omitempty"`
			Quality string `json:"quality,omitempty"`
			URL     string `json:"url,omitempty"`
		} `json:"medias,omitempty"`
	} `json:"video_previews,omitempty"`
	InlineVideoSubtitles []any `json:"inline_video_subtitles,omitempty"`
	CouldPreview         bool  `json:"could_preview,omitempty"`
	VideoCouldPreview    bool  `json:"video_could_preview,omitempty"`
	CoverHidden          bool  `json:"cover_hidden,omitempty"`
	IsRequired           bool  `json:"is_required,omitempty"`
	Extra                struct {
		Rate []struct {
			Type           int  `json:"type,omitempty"`
			CurVersion     int  `json:"cur_version,omitempty"`
			CurRate        int  `json:"cur_rate,omitempty"`
			MaxRate        int  `json:"max_rate,omitempty"`
			TotalRate      int  `json:"total_rate,omitempty"`
			LearnedSeconds int  `json:"learned_seconds,omitempty"`
			IsFinished     bool `json:"is_finished,omitempty"`
		} `json:"rate,omitempty"`
		RatePercent int  `json:"rate_percent,omitempty"`
		IsFinished  bool `json:"is_finished,omitempty"`
		Fav         struct {
			Count   int  `json:"count,omitempty"`
			HadDone bool `json:"had_done,omitempty"`
		} `json:"fav,omitempty"`
		IsUnlocked bool `json:"is_unlocked,omitempty"`
		Learn      struct {
			Ucount int `json:"ucount,omitempty"`
		} `json:"learn,omitempty"`
		FooterCoverData struct {
			ImgURL  string `json:"img_url,omitempty"`
			MpURL   string `json:"mp_url,omitempty"`
			LinkURL string `json:"link_url,omitempty"`
		} `json:"footer_cover_data,omitempty"`
		Work struct {
			IsWork       bool `json:"is_work,omitempty"`
			Status       int  `json:"status,omitempty"`
			CorrectLevel int  `json:"correct_level,omitempty"`
		} `json:"work,omitempty"`
		TypeTime int `json:"type_time,omitempty"`
	} `json:"extra,omitempty"`
	Score           int    `json:"score,omitempty"`
	IsVideo         bool   `json:"is_video,omitempty"`
	PosterWxlite    string `json:"poster_wxlite,omitempty"`
	HadFreelyread   bool   `json:"had_freelyread,omitempty"`
	FloatQrcode     string `json:"float_qrcode,omitempty"`
	FloatAppQrcode  string `json:"float_app_qrcode,omitempty"`
	FloatQrcodeJump string `json:"float_qrcode_jump,omitempty"`
	InPvip          int    `json:"in_pvip,omitempty"`
	CommentCount    int    `json:"comment_count,omitempty"`
	Cshort          string `json:"cshort,omitempty"`
	Like            struct {
		Count   int  `json:"count,omitempty"`
		HadDone bool `json:"had_done,omitempty"`
	} `json:"like,omitempty"`
	ReadingTime int    `json:"reading_time,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	Content     string `json:"content,omitempty"`
	ContentMd   string `json:"content_md,omitempty"`
	Attachments []any  `json:"attachments,omitempty"`
}

type DailyProductRequest struct {
	Type    string `json:"type" form:"type"`
	Size    int    `json:"size" form:"size"`
	Prev    int    `json:"prev" form:"prev"`
	Orderby string `json:"orderby" form:"orderby"`
	LabelID int    `json:"label_id" form:"label_id"`
	Page    int    `json:"-" form:"page"`
	PerPage int    `json:"-"  form:"perPage"`
}

type DailyProductResponse struct {
	Code int `json:"code,omitempty"`
	Data struct {
		Page struct {
			More   bool `json:"more,omitempty"`
			Count  int  `json:"count,omitempty"`
			Score  int  `json:"score,omitempty"`
			Score0 int  `json:"score0,omitempty"`
		} `json:"page,omitempty"`
		List []struct {
			ID        int `json:"id,omitempty"`
			Spu       int `json:"spu,omitempty"`
			Ctime     int `json:"ctime,omitempty"`
			Utime     int `json:"utime,omitempty"`
			BeginTime int `json:"begin_time,omitempty"`
			EndTime   int `json:"end_time,omitempty"`
			Price     struct {
				Market    int `json:"market,omitempty"`
				Sale      int `json:"sale,omitempty"`
				SaleType  int `json:"sale_type,omitempty"`
				StartTime int `json:"start_time,omitempty"`
				EndTime   int `json:"end_time,omitempty"`
			} `json:"price,omitempty"`
			IsSale        bool   `json:"is_sale,omitempty"`
			IsGroupbuy    bool   `json:"is_groupbuy,omitempty"`
			IsPromo       bool   `json:"is_promo,omitempty"`
			IsShareget    bool   `json:"is_shareget,omitempty"`
			IsSharesale   bool   `json:"is_sharesale,omitempty"`
			OnlySellInVip bool   `json:"only_sell_in_vip,omitempty"`
			Type          string `json:"type,omitempty"`
			IsColumn      bool   `json:"is_column,omitempty"`
			IsCore        bool   `json:"is_core,omitempty"`
			IsVideo       bool   `json:"is_video,omitempty"`
			IsAudio       bool   `json:"is_audio,omitempty"`
			IsDailylesson bool   `json:"is_dailylesson,omitempty"`
			IsUniversity  bool   `json:"is_university,omitempty"`
			IsOpencourse  bool   `json:"is_opencourse,omitempty"`
			IsQconp       bool   `json:"is_qconp,omitempty"`
			IsMentor      bool   `json:"is_mentor,omitempty"`
			NavID         int    `json:"nav_id,omitempty"`
			TimeNotSale   int    `json:"time_not_sale,omitempty"`
			Title         string `json:"title,omitempty"`
			Subtitle      string `json:"subtitle,omitempty"`
			Intro         string `json:"intro,omitempty"`
			IntroHTML     string `json:"intro_html,omitempty"`
			Ucode         string `json:"ucode,omitempty"`
			IsFinish      bool   `json:"is_finish,omitempty"`
			Author        struct {
				Name      string `json:"name,omitempty"`
				Intro     string `json:"intro,omitempty"`
				Info      string `json:"info,omitempty"`
				Avatar    string `json:"avatar,omitempty"`
				BriefHTML string `json:"brief_html,omitempty"`
				Brief     string `json:"brief,omitempty"`
				AiID      string `json:"ai_id,omitempty"`
			} `json:"author,omitempty"`
			Cover struct {
				Square            string `json:"square,omitempty"`
				Rectangle         string `json:"rectangle,omitempty"`
				Horizontal        string `json:"horizontal,omitempty"`
				LectureHorizontal string `json:"lecture_horizontal,omitempty"`
				LearnHorizontal   string `json:"learn_horizontal,omitempty"`
				Transparent       string `json:"transparent,omitempty"`
				Color             string `json:"color,omitempty"`
			} `json:"cover,omitempty"`
			Article struct {
				ID                int    `json:"id,omitempty"`
				Count             int    `json:"count,omitempty"`
				CountReq          int    `json:"count_req,omitempty"`
				CountPub          int    `json:"count_pub,omitempty"`
				TotalLength       int    `json:"total_length,omitempty"`
				FirstArticleID    int    `json:"first_article_id,omitempty"`
				FirstArticleTitle string `json:"first_article_title,omitempty"`
			} `json:"article,omitempty"`
			Seo struct {
				Keywords []any `json:"keywords,omitempty"`
			} `json:"seo,omitempty"`
			Share struct {
				Title   string `json:"title,omitempty"`
				Content string `json:"content,omitempty"`
				Cover   string `json:"cover,omitempty"`
				Poster  string `json:"poster,omitempty"`
			} `json:"share,omitempty"`
			Labels     []int  `json:"labels,omitempty"`
			Unit       string `json:"unit,omitempty"`
			ColumnType int    `json:"column_type,omitempty"`
			Column     struct {
				Unit             string `json:"unit,omitempty"`
				Bgcolor          string `json:"bgcolor,omitempty"`
				UpdateFrequency  string `json:"update_frequency,omitempty"`
				IsPreorder       bool   `json:"is_preorder,omitempty"`
				IsFinish         bool   `json:"is_finish,omitempty"`
				IsIncludePreview bool   `json:"is_include_preview,omitempty"`
				ShowChapter      bool   `json:"show_chapter,omitempty"`
				IsSaleProduct    bool   `json:"is_sale_product,omitempty"`
				StudyService     []any  `json:"study_service,omitempty"`
				Path             struct {
					Desc     string `json:"desc,omitempty"`
					DescHTML string `json:"desc_html,omitempty"`
				} `json:"path,omitempty"`
				IsCamp                   bool   `json:"is_camp,omitempty"`
				CatalogPicURL            string `json:"catalog_pic_url,omitempty"`
				RecommendArticles        any    `json:"recommend_articles,omitempty"`
				RecommendComments        any    `json:"recommend_comments,omitempty"`
				Ranks                    any    `json:"ranks,omitempty"`
				HotComments              any    `json:"hot_comments,omitempty"`
				HotLines                 any    `json:"hot_lines,omitempty"`
				DisplayType              int    `json:"display_type,omitempty"`
				IntroBgStyle             int    `json:"intro_bg_style,omitempty"`
				CommentTopAds            string `json:"comment_top_ads,omitempty"`
				ArticleFloatQrcodeURL    string `json:"article_float_qrcode_url,omitempty"`
				ArticleFloatAppQrcodeURL string `json:"article_float_app_qrcode_url,omitempty"`
				ArticleFloatQrcodeJump   string `json:"article_float_qrcode_jump,omitempty"`
				InRank                   bool   `json:"in_rank,omitempty"`
			} `json:"column,omitempty"`
			Dl struct {
				Article struct {
					ID            int    `json:"id,omitempty"`
					VideoDuration string `json:"video_duration,omitempty"`
					VideoHot      int    `json:"video_hot,omitempty"`
					CouldPreview  bool   `json:"could_preview,omitempty"`
				} `json:"article,omitempty"`
				TopicIds []any `json:"topic_ids,omitempty"`
			} `json:"dl,omitempty"`
			University struct {
				TotalHour       int    `json:"total_hour,omitempty"`
				Term            int    `json:"term,omitempty"`
				RedirectType    string `json:"redirect_type,omitempty"`
				RedirectParam   string `json:"redirect_param,omitempty"`
				WxQrcode        string `json:"wx_qrcode,omitempty"`
				WxRule          string `json:"wx_rule,omitempty"`
				ServerStartTime int    `json:"server_start_time,omitempty"`
				LecturerHCover  string `json:"lecturer_h_cover,omitempty"`
				Keywords        string `json:"keywords,omitempty"`
			} `json:"university,omitempty"`
			Opencourse struct {
				VideoBg string `json:"video_bg,omitempty"`
				Ad      struct {
					Cover         string `json:"cover,omitempty"`
					CoverWeb      string `json:"cover_web,omitempty"`
					RedirectType  string `json:"redirect_type,omitempty"`
					RedirectParam string `json:"redirect_param,omitempty"`
				} `json:"ad,omitempty"`
				ArticleFav struct {
					Aid     int  `json:"aid,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
					Count   int  `json:"count,omitempty"`
				} `json:"article_fav,omitempty"`
				AuthorHCover string `json:"author_h_cover,omitempty"`
			} `json:"opencourse,omitempty"`
			Qconp struct {
				TopicID      int    `json:"topic_id,omitempty"`
				CoverAppoint string `json:"cover_appoint,omitempty"`
				Article      struct {
					ID            int    `json:"id,omitempty"`
					Cover         string `json:"cover,omitempty"`
					VideoDuration string `json:"video_duration,omitempty"`
					VideoHot      int    `json:"video_hot,omitempty"`
				} `json:"article,omitempty"`
			} `json:"qconp,omitempty"`
			FavQrcode string `json:"fav_qrcode,omitempty"`
			Extra     struct {
				Sub struct {
					Count      int  `json:"count,omitempty"`
					HadDone    bool `json:"had_done,omitempty"`
					CouldOrder bool `json:"could_order,omitempty"`
					AccessMask int  `json:"access_mask,omitempty"`
				} `json:"sub,omitempty"`
				Fav struct {
					Count   int  `json:"count,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
				} `json:"fav,omitempty"`
				Rate struct {
					ArticleCount    int  `json:"article_count,omitempty"`
					ArticleCountReq int  `json:"article_count_req,omitempty"`
					IsFinished      bool `json:"is_finished,omitempty"`
					RatePercent     int  `json:"rate_percent,omitempty"`
					VideoSeconds    int  `json:"video_seconds,omitempty"`
					LastArticleID   int  `json:"last_article_id,omitempty"`
					LastChapterID   int  `json:"last_chapter_id,omitempty"`
					HasLearn        bool `json:"has_learn,omitempty"`
				} `json:"rate,omitempty"`
				Cert struct {
					ID   string `json:"id,omitempty"`
					Type int    `json:"type,omitempty"`
				} `json:"cert,omitempty"`
				Nps struct {
					Min    int    `json:"min,omitempty"`
					Status int    `json:"status,omitempty"`
					URL    string `json:"url,omitempty"`
				} `json:"nps,omitempty"`
				AnyRead struct {
					Total int `json:"total,omitempty"`
					Count int `json:"count,omitempty"`
				} `json:"any_read,omitempty"`
				University struct {
					Status               int    `json:"status,omitempty"`
					ViewStatus           int    `json:"view_status,omitempty"`
					ChargeStatus         int    `json:"charge_status,omitempty"`
					ShareRenewalStatus   int    `json:"share_renewal_status,omitempty"`
					UnlockedStatus       int    `json:"unlocked_status,omitempty"`
					UnlockedChapterIds   []any  `json:"unlocked_chapter_ids,omitempty"`
					UnlockedChapterID    int    `json:"unlocked_chapter_id,omitempty"`
					UnlockedChapterTitle string `json:"unlocked_chapter_title,omitempty"`
					UnlockedArticleCount int    `json:"unlocked_article_count,omitempty"`
					UnlockedNextTime     int    `json:"unlocked_next_time,omitempty"`
					ExpireTime           int    `json:"expire_time,omitempty"`
					IsExpired            bool   `json:"is_expired,omitempty"`
					IsGraduated          bool   `json:"is_graduated,omitempty"`
					HadSub               bool   `json:"had_sub,omitempty"`
					Timeline             []any  `json:"timeline,omitempty"`
					HasWxFriend          bool   `json:"has_wx_friend,omitempty"`
					StartTime            int    `json:"start_time,omitempty"`
					SubTermTitle         string `json:"sub_term_title,omitempty"`
					SubSku               int    `json:"sub_sku,omitempty"`
				} `json:"university,omitempty"`
				Vip struct {
					IsYearCard bool `json:"is_year_card,omitempty"`
					Show       bool `json:"show,omitempty"`
					EndTime    int  `json:"end_time,omitempty"`
				} `json:"vip,omitempty"`
				Appoint struct {
					CouldDo bool `json:"could_do,omitempty"`
					HadDone bool `json:"had_done,omitempty"`
					Count   int  `json:"count,omitempty"`
				} `json:"appoint,omitempty"`
				GroupBuy struct {
					SuccessUcount int    `json:"success_ucount,omitempty"`
					JoinCode      string `json:"join_code,omitempty"`
					CouldGroupbuy bool   `json:"could_groupbuy,omitempty"`
					HadJoin       bool   `json:"had_join,omitempty"`
					Price         int    `json:"price,omitempty"`
					List          []any  `json:"list,omitempty"`
				} `json:"group_buy,omitempty"`
				Sharesale struct {
					OriginalPicColor    string `json:"original_pic_color,omitempty"`
					OriginalPicURL      string `json:"original_pic_url,omitempty"`
					PromoPicColor       string `json:"promo_pic_color,omitempty"`
					PromoPicURL         string `json:"promo_pic_url,omitempty"`
					ShareSalePrice      int    `json:"share_sale_price,omitempty"`
					ShareSaleGuestPrice int    `json:"share_sale_guest_price,omitempty"`
				} `json:"sharesale,omitempty"`
				Promo struct {
					EntTime int `json:"ent_time,omitempty"`
				} `json:"promo,omitempty"`
				Channel struct {
					Is         bool `json:"is,omitempty"`
					BackAmount int  `json:"back_amount,omitempty"`
				} `json:"channel,omitempty"`
				FirstPromo struct {
					Price     int  `json:"price,omitempty"`
					CouldJoin bool `json:"could_join,omitempty"`
				} `json:"first_promo,omitempty"`
				CouponPromo struct {
					CouldJoin bool `json:"could_join,omitempty"`
					Price     int  `json:"price,omitempty"`
				} `json:"coupon_promo,omitempty"`
				Helper []any `json:"helper,omitempty"`
				Tab    struct {
					Comment bool `json:"comment,omitempty"`
					Package bool `json:"package,omitempty"`
				} `json:"tab,omitempty"`
				Modules   []any `json:"modules,omitempty"`
				Cid       int   `json:"cid,omitempty"`
				FirstAids []any `json:"first_aids,omitempty"`
				StudyPlan struct {
					ID              int `json:"id,omitempty"`
					DayNums         int `json:"day_nums,omitempty"`
					ArticleNums     int `json:"article_nums,omitempty"`
					LearnedWeekNums int `json:"learned_week_nums,omitempty"`
					Status          int `json:"status,omitempty"`
				} `json:"study_plan,omitempty"`
				CateID   int    `json:"cate_id,omitempty"`
				CateName string `json:"cate_name,omitempty"`
				GroupTag struct {
					IsRecommend     bool `json:"is_recommend,omitempty"`
					IsRecentlyLearn bool `json:"is_recently_learn,omitempty"`
					IsTop           bool `json:"is_top,omitempty"`
				} `json:"group_tag,omitempty"`
				FirstAward struct {
					Show          bool   `json:"show,omitempty"`
					Talks         int    `json:"talks,omitempty"`
					Reads         int    `json:"reads,omitempty"`
					Amount        int    `json:"amount,omitempty"`
					ExpireTime    int    `json:"expire_time,omitempty"`
					RedirectType  string `json:"redirect_type,omitempty"`
					RedirectParam string `json:"redirect_param,omitempty"`
				} `json:"first_award,omitempty"`
				VipPromo struct {
					DiscountLevel int `json:"discount_level,omitempty"`
					DiscountPrice int `json:"discount_price,omitempty"`
					MinLevel      int `json:"min_level,omitempty"`
					Rules         any `json:"rules,omitempty"`
				} `json:"vip_promo,omitempty"`
				IsTgoTicket bool `json:"is_tgo_ticket,omitempty"`
				BPack       any  `json:"b_pack,omitempty"`
				PSkus       any  `json:"p_skus,omitempty"`
			} `json:"extra,omitempty"`
			AvailableCoupons any    `json:"available_coupons,omitempty"`
			InPvip           int    `json:"in_pvip,omitempty"`
			IsJoinCvip       int    `json:"is_join_cvip,omitempty"`
			ColumnBadge      string `json:"column_badge,omitempty"`
			HideCopyright    bool   `json:"hide_copyright,omitempty"`
		} `json:"list,omitempty"`
		Topics    []any `json:"topics,omitempty"`
		Articles  []any `json:"articles,omitempty"`
		LookLists []any `json:"look_lists,omitempty"`
	} `json:"data,omitempty"`
	Error any `json:"error,omitempty"`
	Extra struct {
		Cost      float64 `json:"cost,omitempty"`
		RequestID any     `json:"request-id,omitempty"`
	} `json:"extra,omitempty"`
}
