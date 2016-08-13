## api服务
[![Build Status](https://drone.io/github.com/btlike/api/status.png)](https://drone.io/github.com/btlike/api/latest)

### 基础接口

- 根据关键词搜索
- 根据infohash获取详情(文件大小、文件名、文件数量等)
- 热门排行
- 首页推荐

### 安装
`go get github.com/btlike/api`



### 接口说明

- 获取首页推荐列表

  - 请求

    ```http
    GET /recommend
    ```

  - 返回

    ```json
    [
      {
        "Id": 1,
        "Name": "疯狂动物城"
      },
      {
        "Id": 2,
        "Name": "星际穿越"
      }
    ]
    ```

- 获取热门排行（最近一月）

  - 请求

    ```http
    GET /trend?type=month
    ```

  - 返回

    ```json
    [
      {
        "Name": "Vikk.I.OVTsa.2O16.D.WEB-DLRip.1400MB.avi",
        "ID": "44F0F0BE6E0A48AEA019F2D5D3F1F7B6214EEA0A",
        "Heat": 11731,
        "Length": 1467510784,
        "CreateTime": "2016-07-14T00:13:43.957687015Z"
      },
      {
        "Name": "snis702.avi",
        "ID": "4EC957F7946016800E1D83C391CE8D1B09F9A182",
        "Heat": 11458,
        "Length": 1500338387,
        "CreateTime": "2016-07-15T17:42:30.329504348Z"
      }
    ]  
    ```

- 获取热门排行（最近一周）

  - 请求

    ```http
    GET /trend?type=week
    ```

  - 返回

    ```json
    [
      {
        "Name": "Vikk.I.OVTsa.2O16.D.WEB-DLRip.1400MB.avi",
        "ID": "44F0F0BE6E0A48AEA019F2D5D3F1F7B6214EEA0A",
        "Heat": 11731,
        "Length": 1467510784,
        "CreateTime": "2016-07-14T00:13:43.957687015Z"
      },
      {
        "Name": "snis702.avi",
        "ID": "4EC957F7946016800E1D83C391CE8D1B09F9A182",
        "Heat": 11458,
        "Length": 1500338387,
        "CreateTime": "2016-07-15T17:42:30.329504348Z"
      }
    ]  
    ```

- 搜索关键词

  - 请求

    ```http
    GET /list?keyword=something&page=1&order=h

    keyword：关键词
    page：页码
    order：排序(l:时间排序,m:大小排序,h:热度排序,x:相关度排序)
    ```

  - 返回

    ```json
    {
      "Torrent": [
        {
          "Infohash": "750160421F636C745DA81904D9BE5E64581540D0",
          "Name": "星际穿越.avi",
          "CreateTime": "2016-07-21T19:36:36.041734298Z",
          "Length": 2793154276,
          "FileCount": 1,
          "Heat": 1,
          "Files": [
            {
              "Name": "星际穿越.avi",
              "Length": 2793154276
            }
          ]
        },
        {
          "Infohash": "8616C437C41CD6E13203E6F1E24BE3B3AF77FDF1",
          "Name": "[www.mitbbs.xyz]星际穿越Interstellar.mp4",
          "CreateTime": "2016-07-03T22:52:57.998444942Z",
          "Length": 2071404986,
          "FileCount": 1,
          "Heat": 6,
          "Files": [
            {
              "Name": "[www.mitbbs.xyz]星际穿越Interstellar.mp4",
              "Length": 2071404986
            }
          ]
        }
      ],
      "Count": 33
    }
    ```

- 查询资源详情

  - 请求

    ```http
    GET /detail?id=215ED94A26E41DB6E5DA945F57B0465F0987A734
    ```

  - 返回

    ```json
    {
      "Infohash": "215ED94A26E41DB6E5DA945F57B0465F0987A734",
      "Name": "Theres Something About Mary EXTENDED (1998)",
      "CreateTime": "2016-06-21T05:00:57.905703416+08:00",
      "Length": 893394668,
      "FileCount": 3,
      "Heat": 0,
      "Files": [
        {
          "Name": "Theres.Something.About.Mary.EXTENDED.CUT.1998.720p.BrRip.x264.YIFY.mp4",
          "Length": 893121852
        },
        {
          "Name": "Theres.Something.About.Mary.EXTENDED.CUT.1998.720p.BrRip.x264.YIFY.srt",
          "Length": 142139
        },
        {
          "Name": "WWW.YIFY-TORRENTS.COM.jpg",
          "Length": 130677
        }
      ]
    }
    ```

    ​

  ​
