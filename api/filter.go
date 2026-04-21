package api

import (
	"pansou/model"
	"strings"
)

// applyResultFilter 应用过滤器到搜索响应
func applyResultFilter(response model.SearchResponse, filter *model.FilterConfig, resultType string) model.SearchResponse {
	if filter == nil || (len(filter.Include) == 0 && len(filter.Exclude) == 0) {
		return response
	}

	// 预处理关键词（转小写）
	includeKeywords := make([]string, len(filter.Include))
	for i, kw := range filter.Include {
		includeKeywords[i] = strings.ToLower(kw)
	}
	
	excludeKeywords := make([]string, len(filter.Exclude))
	for i, kw := range filter.Exclude {
		excludeKeywords[i] = strings.ToLower(kw)
	}

	// 根据结果类型决定过滤策略
	if resultType == "merged_by_type" || resultType == "" {
		// 过滤 merged_by_type 的 note 字段
		response.MergedByType = filterMergedByType(response.MergedByType, includeKeywords, excludeKeywords)
		
		// 重新计算 total
		total := 0
		for _, links := range response.MergedByType {
			total += len(links)
		}
		response.Total = total
	} else if resultType == "all" || resultType == "results" {
		// 过滤 results 的 title 和 links 的 work_title
		response.Results = filterResults(response.Results, includeKeywords, excludeKeywords)
		response.Total = len(response.Results)
		
		// 如果是 all 类型，也需要过滤 merged_by_type
		if resultType == "all" {
			response.MergedByType = filterMergedByType(response.MergedByType, includeKeywords, excludeKeywords)
		}
	}

	return response
}

// filterMergedByType 过滤 merged_by_type 中的链接
func filterMergedByType(mergedLinks model.MergedLinks, includeKeywords, excludeKeywords []string) model.MergedLinks {
	if mergedLinks == nil {
		return nil
	}

	filtered := make(model.MergedLinks)
	
	for linkType, links := range mergedLinks {
		filteredLinks := make([]model.MergedLink, 0)
		
		for _, link := range links {
			if matchFilter(link.Note, includeKeywords, excludeKeywords) {
				filteredLinks = append(filteredLinks, link)
			}
		}
		
		// 只添加非空的类型
		if len(filteredLinks) > 0 {
			filtered[linkType] = filteredLinks
		}
	}
	
	return filtered
}

// filterResults 过滤 results 数组
func filterResults(results []model.SearchResult, includeKeywords, excludeKeywords []string) []model.SearchResult {
	if results == nil {
		return nil
	}

	filtered := make([]model.SearchResult, 0)
	
	for _, result := range results {
		// 先检查 title 是否匹配
		if !matchFilter(result.Title, includeKeywords, excludeKeywords) {
			continue
		}
		
		// title 匹配后，过滤 links 中的 work_title
		filteredLinks := make([]model.Link, 0)
		for _, link := range result.Links {
			// 如果 link 有 work_title，检查它；否则使用 result.Title
			checkText := link.WorkTitle
			if checkText == "" {
				checkText = result.Title
			}
			
			if matchFilter(checkText, includeKeywords, excludeKeywords) {
				filteredLinks = append(filteredLinks, link)
			}
		}
		
		// 只有有链接的结果才添加
		if len(filteredLinks) > 0 {
			result.Links = filteredLinks
			filtered = append(filtered, result)
		}
	}
	
	return filtered
}

// matchFilter 检查文本是否匹配过滤条件
// 规则：exclude 优先于 include；若文本包含任意 exclude 关键词则排除，
// 若设置了 include 则文本必须包含至少一个 include 关键词才保留。
func matchFilter(text string, includeKeywords, excludeKeywords []string) bool {
	lowerText := strings.ToLower(text)

	// 先检查排除关键词（优先级更高）
	for _, kw := range excludeKeywords {
		if strings.Contains(lowerText, kw) {
			return false
		}
	}

	// 再检查包含关键词
	if len(includeKeywords) > 0 {
		for _, kw := range includeKeywords {
			if strings.Contains(lowerText, kw) {
				return true
			}
		}
		return false
	}

	return true
}
