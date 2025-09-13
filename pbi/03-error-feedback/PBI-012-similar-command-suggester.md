# PBI-012: 類似コマンド提案（Levenshtein距離）

## 概要
存在しないコマンド・サブコマンドに対して、Levenshtein距離アルゴリズムを使用して類似のコマンドを検索し、ユーザーに建設的な候補を提案する機能を実装する。typoや記憶違いによるエラーをユーザーフレンドリーに解決する。

## 受け入れ条件 ✅ **完了 2025-01-09**
- [x] Levenshtein距離アルゴリズムが正確に実装されている
- [x] メインコマンド・サブコマンドの両方で類似検索が動作する
- [x] 適切な閾値で候補を絞り込んでいる
- [x] 候補の優先順位付けが適切に行われている
- [x] パフォーマンスが実用的なレベルにある

## 技術仕様

### Levenshtein距離アルゴリズム

#### 概要
2つの文字列間の編集距離（挿入・削除・置換の最小回数）を計算するアルゴリズム。typo検出に最適。

#### 実装例
```go
// internal/validation/similar_command_suggester.go
package validation

import (
    "sort"
    "strings"
)

// SimilarityResult は類似性検索結果
type SimilarityResult struct {
    Command  string  // 候補コマンド
    Distance int     // Levenshtein距離
    Score    float64 // 類似度スコア（0.0-1.0）
}

// SimilarCommandSuggester は類似コマンド提案器
type SimilarCommandSuggester struct {
    allCommands       []string            // 全コマンドリスト
    commandSubcommands map[string][]string // コマンド->サブコマンドマッピング
    maxDistance       int                 // 最大許容距離
    maxSuggestions    int                 // 最大提案数
}

// NewSimilarCommandSuggester は新しい提案器を作成
func NewSimilarCommandSuggester(maxDistance, maxSuggestions int) *SimilarCommandSuggester {
    return &SimilarCommandSuggester{
        allCommands:    getAllCommands(),
        commandSubcommands: getAllCommandSubcommands(),
        maxDistance:    maxDistance,
        maxSuggestions: maxSuggestions,
    }
}

// LevenshteinDistance はLevenshtein距離を計算
func (s *SimilarCommandSuggester) LevenshteinDistance(s1, s2 string) int {
    s1 = strings.ToLower(s1)
    s2 = strings.ToLower(s2)
    
    if s1 == s2 {
        return 0
    }
    
    if len(s1) == 0 {
        return len(s2)
    }
    
    if len(s2) == 0 {
        return len(s1)
    }
    
    // 動的プログラミングによる実装
    matrix := make([][]int, len(s1)+1)
    for i := range matrix {
        matrix[i] = make([]int, len(s2)+1)
    }
    
    // 初期化
    for i := 0; i <= len(s1); i++ {
        matrix[i][0] = i
    }
    for j := 0; j <= len(s2); j++ {
        matrix[0][j] = j
    }
    
    // 距離計算
    for i := 1; i <= len(s1); i++ {
        for j := 1; j <= len(s2); j++ {
            cost := 0
            if s1[i-1] != s2[j-1] {
                cost = 1
            }
            
            matrix[i][j] = min(
                matrix[i-1][j]+1,     // 削除
                matrix[i][j-1]+1,     // 挿入
                matrix[i-1][j-1]+cost, // 置換
            )
        }
    }
    
    return matrix[len(s1)][len(s2)]
}
```

### コマンド提案機能
```go
// SuggestMainCommands はメインコマンドの候補を提案
func (s *SimilarCommandSuggester) SuggestMainCommands(input string) []SimilarityResult {
    var results []SimilarityResult
    
    for _, command := range s.allCommands {
        distance := s.LevenshteinDistance(input, command)
        
        if distance <= s.maxDistance {
            score := 1.0 - float64(distance)/float64(max(len(input), len(command)))
            results = append(results, SimilarityResult{
                Command:  command,
                Distance: distance,
                Score:    score,
            })
        }
    }
    
    // スコア順でソート
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    
    // 最大提案数に制限
    if len(results) > s.maxSuggestions {
        results = results[:s.maxSuggestions]
    }
    
    return results
}

// SuggestSubcommands はサブコマンドの候補を提案
func (s *SimilarCommandSuggester) SuggestSubcommands(mainCommand, input string) []SimilarityResult {
    subcommands, exists := s.commandSubcommands[mainCommand]
    if !exists {
        return nil
    }
    
    var results []SimilarityResult
    
    for _, subcommand := range subcommands {
        distance := s.LevenshteinDistance(input, subcommand)
        
        if distance <= s.maxDistance {
            score := 1.0 - float64(distance)/float64(max(len(input), len(subcommand)))
            results = append(results, SimilarityResult{
                Command:  subcommand,
                Distance: distance,
                Score:    score,
            })
        }
    }
    
    // スコア順でソート
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    
    // 最大提案数に制限
    if len(results) > s.maxSuggestions {
        results = results[:s.maxSuggestions]
    }
    
    return results
}
```

### 特別な考慮事項

#### 1. 閾値の調整
```go
// 推奨設定
const (
    DefaultMaxDistance    = 3  // 最大3文字の違いまで
    DefaultMaxSuggestions = 5  // 最大5個の候補
    MinScore             = 0.5 // 最低類似度50%
)

// 動的閾値（入力文字列長に応じて調整）
func (s *SimilarCommandSuggester) getAdaptiveMaxDistance(input string) int {
    length := len(input)
    switch {
    case length <= 3:
        return 1  // 短い文字列は厳しく
    case length <= 6:
        return 2  // 中程度の文字列
    default:
        return 3  // 長い文字列は緩く
    }
}
```

#### 2. 一般的なtypoパターンの重み付け
```go
// CommonTypoPatterns は一般的なtypoパターン
var CommonTypoPatterns = map[string][]string{
    "server":   {"sever", "serv", "srv", "servers"},
    "disk":     {"disc", "dsk", "disks"},
    "database": {"db", "databse", "datbase"},
    "list":     {"lst", "lis"},
    "create":   {"creat", "crate"},
    "delete":   {"delet", "del"},
}

// getTypoScore はtypoパターンに基づく追加スコアを計算
func (s *SimilarCommandSuggester) getTypoScore(input, candidate string) float64 {
    patterns, exists := CommonTypoPatterns[candidate]
    if !exists {
        return 0.0
    }
    
    for _, pattern := range patterns {
        if strings.ToLower(input) == strings.ToLower(pattern) {
            return 0.2 // typoパターンマッチには追加スコア
        }
    }
    
    return 0.0
}
```

### パフォーマンス最適化
```go
// キャッシュ機能
type suggestionCache struct {
    cache map[string][]SimilarityResult
    maxEntries int
}

// プリフィックス検索による初期絞り込み
func (s *SimilarCommandSuggester) filterByPrefix(input string, candidates []string) []string {
    if len(input) < 2 {
        return candidates // 短すぎる場合は全候補を対象
    }
    
    prefix := strings.ToLower(input[:2])
    var filtered []string
    
    for _, candidate := range candidates {
        if strings.HasPrefix(strings.ToLower(candidate), prefix) {
            filtered = append(filtered, candidate)
        }
    }
    
    // プリフィックスマッチがない場合は全候補を対象
    if len(filtered) == 0 {
        return candidates
    }
    
    return filtered
}
```

## テスト戦略
- アルゴリズムテスト：Levenshtein距離の計算が正確であることを確認
- typoテスト：一般的なtypoパターンが適切に検出されることを確認
- 閾値テスト：様々な閾値設定で適切な候補が提案されることを確認
- パフォーマンステスト：大量のコマンドに対して実用的な速度で動作することを確認
- 境界値テスト：極端に短い/長い入力に対して適切に動作することを確認
- 統合テスト：実際のコマンド辞書を使用して現実的な候補が提案されることを確認

## 依存関係
- 前提PBI: PBI-001～006 (コマンド辞書) - 候補検索の対象データ
- 関連PBI: PBI-011 (エラーメッセージ生成) - 候補をメッセージに組み込む

## 見積もり
- 開発工数: 5時間
  - Levenshtein距離アルゴリズム実装: 2時間
  - 候補検索・ランキング機能: 2時間
  - パフォーマンス最適化: 0.5時間
  - ユニットテスト作成: 0.5時間

## 完了の定義 ✅ **完了 2025-01-09**
- [x] `internal/validation/similar_command_suggester.go`ファイルが作成されている
- [x] `SimilarCommandSuggester`構造体と提案メソッドが実装されている
- [x] Levenshtein距離アルゴリズムが正確に実装されている
- [x] メインコマンド・サブコマンドの候補提案が正しく動作する
- [x] 適切な閾値とランキングで候補が絞り込まれる
- [x] typoパターンの特別処理が実装されている
- [x] パフォーマンスが実用的なレベルにある（1000コマンドを1秒以内で処理）
- [x] 包括的なユニットテストが作成され、すべて通過している
- [x] 実際のコマンド辞書を使用したテストが通過している
- [x] コードレビューが完了している

## 実装結果 📊

**実装ファイル:**
- `internal/validation/similar_command_suggester.go` - SimilarCommandSuggester構造体とLevenshtein距離アルゴリズムの完全実装
- `internal/validation/similar_command_suggester_test.go` - 20テスト関数による包括的テスト

**実装内容:**
- 動的プログラミングによるLevenshtein距離アルゴリズム実装
- メインコマンド・サブコマンド両方の類似コマンド提案機能
- 適応的な距離閾値（入力文字列長に応じて1-3文字の差まで許容）
- 11の一般的typoパターン辞書（server→sever, disk→disc等）
- プリフィックスフィルタリングによるパフォーマンス最適化
- スコアベースの候補ランキング（0.0-1.0の類似度スコア）
- typoパターンマッチ時の追加ボーナス（+0.2スコア）
- 最大提案数制限機能（デフォルト5個まで）

**テスト結果:**
- 20のテスト関数すべて成功
- Levenshtein距離計算の正確性検証
- メインコマンド・サブコマンド提案機能の検証
- 適応的閾値ロジックの検証
- typoパターン認識の検証
- エッジケース処理の検証（短い入力、長い入力、特殊文字等）
- パフォーマンステスト（実用的速度での動作確認）

**技術的特徴:**
- 効率的な動的プログラミングによるO(m*n)時間計算量
- プリフィックスフィルタリングによる候補数削減
- 関数名重複回避のための独自min/max関数実装
- スコア上限制御による0.0-1.0範囲の保証
- 既存コマンド辞書との完全統合
- 拡張可能なtypoパターン管理システム

## 備考
- この機能はユーザビリティに大きく影響する重要な機能
- 正確性とパフォーマンスのバランスが重要
- 一般的なtypoパターンの理解がユーザーエクスペリエンスを向上させる
- 将来的なコマンド追加に対応できる拡張可能な設計