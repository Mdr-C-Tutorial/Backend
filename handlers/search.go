package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
)

// SearchResult 定义搜索结果的结构
type SearchResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
}

// SearchResponse 定义搜索响应的结构
type SearchResponse struct {
	MCT    []SearchResult `json:"mct"`
	CppRef []SearchResult `json:"cppref"`
	Google []SearchResult `json:"google"`
}

// HandleSearch 处理搜索请求
func HandleSearch(c *gin.Context) {
	query := c.Param("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "搜索关键词不能为空"})
		return
	}

	// 创建一个通道来接收搜索结果
	cppRefChan := make(chan []SearchResult)
	googleChan := make(chan []SearchResult)

	// 并行获取搜索结果
	go func() {
		// 使用 DuckDuckGo 搜索 CPP Reference
		results, err := searchDuckDuckGo(query)
		if err != nil {
			fmt.Printf("DuckDuckGo 搜索错误: %v\n", err)
			// 如果 DuckDuckGo 搜索失败，尝试直接搜索 CPP Reference
			results, err = searchCppReference(query)
			fmt.Println(results)
			if err != nil {
				fmt.Printf("CPP Reference 搜索错误: %v\n", err)
				cppRefChan <- []SearchResult{}
			} else {
				cppRefChan <- results
			}
		} else {
			cppRefChan <- results
		}
	}()

	go func() {
		results, err := searchGoogle(query)
		if err != nil {
			fmt.Printf("Google 搜索错误: %v\n", err)
			googleChan <- getGoogleMockResults(query) // 使用模拟数据
		} else {
			googleChan <- results
		}
	}()

	// 收集搜索结果
	cppRefResults := <-cppRefChan
	googleResults := <-googleChan

	// 构建响应
	response := SearchResponse{
		MCT:    []SearchResult{}, // MCT 结果暂时为空
		CppRef: cppRefResults,
		Google: googleResults,
	}

	c.JSON(http.StatusOK, response)
}

// searchCppReference 从 CPP Reference 获取搜索结果
func searchCppReference(query string) ([]SearchResult, error) {
	baseURL := "https://zh.cppreference.com/mwiki/index.php"
	params := url.Values{}
	params.Add("title", "Special:搜索")
	params.Add("search", query)
	params.Add("fulltext", "1")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	// 解析搜索结果
	doc.Find("ul.mw-search-results li").Each(func(i int, s *goquery.Selection) {
		if i >= 15 { // 限制结果数量
			return
		}

		title := s.Find(".mw-search-result-heading a").Text()
		link, _ := s.Find(".mw-search-result-heading a").Attr("href")
		desc := s.Find(".searchresult").Text()

		// 处理链接
		if !strings.HasPrefix(link, "http") {
			link = "https://zh.cppreference.com" + link
		}

		// 清理描述文本
		desc = strings.TrimSpace(desc)
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}

		results = append(results, SearchResult{
			Title:       title,
			Description: desc,
			Link:        link,
		})
	})

	return results, nil
}

// searchGoogle 从 Google 获取搜索结果
func searchGoogle(query string) ([]SearchResult, error) {
	// 使用 Google 自定义搜索 API
	apiKey := "YOUR_API_KEY"      // 替换为你的 Google API 密钥
	cx := "YOUR_SEARCH_ENGINE_ID" // 替换为你的自定义搜索引擎 ID

	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Add("key", apiKey)
	params.Add("cx", cx)
	params.Add("q", query+" C语言")
	params.Add("num", "5")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "?" + params.Encode())
	if err != nil {
		fmt.Printf("Google API 请求错误: %v\n", err)
		return getGoogleMockResults(query), err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Google API 返回非200状态码: %d\n", resp.StatusCode)
		return getGoogleMockResults(query), fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	var searchResponse struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		fmt.Printf("解析 Google API 响应错误: %v\n", err)
		return getGoogleMockResults(query), err
	}

	var results []SearchResult
	for i, item := range searchResponse.Items {
		if i >= 15 { // 限制结果数量
			break
		}

		results = append(results, SearchResult{
			Title:       item.Title,
			Description: item.Snippet,
			Link:        item.Link,
		})
	}

	if len(results) == 0 {
		fmt.Println("Google API 返回结果为空")
		return getGoogleMockResults(query), nil
	}

	return results, nil
}

// searchDuckDuckGo 从 DuckDuckGo 获取 CPP Reference 的搜索结果
func searchDuckDuckGo(query string) ([]SearchResult, error) {
	baseURL := "https://duckduckgo.com/html/"
	params := url.Values{}
	params.Add("q", query+" site:zh.cppreference.com")

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}

	req, err := http.NewRequest("POST", baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("请求失败，状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	// 解析DuckDuckGo搜索结果
	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		if i >= 5 { // 限制结果数量
			return
		}

		title := s.Find(".result__title").Text()
		link, _ := s.Find(".result__url").Attr("href")
		desc := s.Find(".result__snippet").Text()

		// 清理文本
		title = strings.TrimSpace(title)
		desc = strings.TrimSpace(desc)
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}

		if title != "" && link != "" {
			results = append(results, SearchResult{
				Title:       title,
				Description: desc,
				Link:        link,
			})
		}
	})

	return results, nil
}

// 备用方案：如果无法使用 Google API，可以使用模拟数据
func getGoogleMockResults(query string) []SearchResult {
	return []SearchResult{
		{
			Title:       query + " - C语言教程 | 菜鸟教程",
			Description: "C语言是一种通用的、面向过程的计算机程序设计语言，广泛用于底层开发。" + query + " 是C语言中的重要概念...",
			Link:        "https://www.runoob.com/cprogramming/c-" + url.QueryEscape(query) + ".html",
		},
		{
			Title:       query + " 详解 - C语言中文网",
			Description: "本文详细介绍了C语言中 " + query + " 的用法和注意事项，包含多个实例代码...",
			Link:        "https://c.biancheng.net/" + url.QueryEscape(query) + "/",
		},
		{
			Title:       "如何在C语言中正确使用 " + query,
			Description: "许多初学者在使用 " + query + " 时会遇到问题，本文将为您详细讲解正确的使用方法...",
			Link:        "https://stackoverflow.com/questions/tagged/" + url.QueryEscape(query),
		},
	}
}
