### SafeSQLが面白い
- Disclaimer: 雑メモだから間違ってるかも
#### SQL Injectionを防ぎたい
- 1. 外からの入力をSQL文の組み立てに使わない
- 2. 外からの入力を渡す場合はPlaceholderを使う
```go
db.Query("SELECT ... WHERE size = ?", 10)
```

#### SafeSQL が面白い
- GoogleはSacure by defaultの考え方で作られたライブラリを公開している
  - https://github.com/google/go-safeweb
  - https://github.com/google/tink

- go-safewebに入ってるSafeSQLが面白かった
  - https://github.com/google/go-safeweb/blob/master/safesql/safesql.go
  - Secure by defaultで、安全な状態を強制する. 間違えなくする

#### 仕組み
- 別パッケージのunexportedなtype stringConstant
  - New(text stringConstant)は、文字列リテラル("hoge")であれば、呼び出せる
  - string型の変数を渡そうとすると、コンパイル時に弾かれる(unexportedなので型キャストもできない)
    - Runtime Errorでなく、コンパイル時にエラーが出る！
```go
type stringConstant string

func New(text stringConstant) TrustedSQLString { return TrustedSQLString{string(text)} }
```

#### 使い方
- https://github.com/rung/public-note/blob/main/safesql/main.go
```go
// database/sqlのwrapper
var db  *safesql.DB
---
	// SafeSQLを用いた呼び出し (pass/コンパイルできる)
	age := 1
	db.Query(safesql.New("SELECT hoge FROM hoge WHERE age = ?"), age)

	// SafeSQL (連結) (pass/コンパイルできる)
	s1 := safesql.New("SELECT hoge")
	s2 := safesql.New(" FROM hoge WHERE age = ?")

	db.Query(safesql.TrustedSQLStringConcat(s1, s2), age)

	// SafeSQL (fail/コンパイルエラー)
	//  Compile時に文字列リテラルであれば、safesql.Newが使えるが、string型を渡す形だと使えない
	s3fromExternal := safesql.New(os.Getenv("HOGE"))
```
- 外からの入力を防ぐのを、Runtime Errorでなく、Compile Errorで防げる...!
- コンパイル時に決定できるのであればコンパイルできるし、できないのであればできない

#### SQL Injectionを防ぐ静的解析系ツールとの比較 (うろおぼえ!)
- [stripe/safesql](https://github.com/stripe/safesql)
  - 基本的にはdb.queryに入る文字列の連結を防ぐアプローチ
  - safesqlはSSAを用いてdb.queryに入れる文字列を検知して、いい感じだが、別の関数で連結された文字列を、db.queryにわたす場合は検知できない

- gosec
  - gosecは誤検知や検知漏れが発生する（[SQLっぽい文字列連結の検知をする](https://github.com/securego/gosec/blob/e3dffd64501211e83308009841047d9c8c4964d2/rules/sql.go#L128). パターンに当てはまらなかったら検知しない）

#### Safesqlを静的解析で拡張する(Google)
- Building Secure and Reliable Systemsで紹介されてる
- https://github.com/google/go-safeweb にも説明あり

- アプローチ
  - 基本はSafeSQLを経由してのDB呼び出しを強制する (Secure By Defaultのアプローチで、基本が安全な状態にする)
    - 静的解析によって、safesql使ってなかったらreportingしてSecurity TeamのReviewをいれたり、強制的に使わせたりするアプローチがとれる
    - 危ないことをさせるときは、危ないことが分かるようにする
      - GoのUnsafeも同じような話

- [Bulding Secure and Reliable Systems](https://sre.google/books/building-secure-reliable-systems/) (Google SRE本3冊目) "Chapter 12. Writing Code" からの引用
  - In theory, you can create secure and reliable software by carefully writing application code that maintains these invariants. However, as the number of desired properties and the size of the codebase grows, this approach becomes almost impossible. It’s unreasonable to expect any developer to be an expert in all these subjects, or to constantly maintain vigilance when writing or reviewing code.
    - (拙訳) 理論的には、アプリケーションがセキュアで信頼性があるような不変条件を維持しつつ注意深くコードを書くことで、セキュアで信頼性のあるソフトウェアを作ることができます。 しかし、そのために望まれる要素の数や、コードベースの増加とともに、このアプローチはほとんど不可能になります。全ての開発者がこれらの話についてエキスパートであったり、コードを書いたりレビューするたびに注意を払うというのは、非合理的な想定なのです。
  - In our experience, as long as the feedback cycle is quick, and fixing each pattern is relatively easy, developers embrace inherently safe APIs much more readily—even when we can’t prove that their code was insecure, or when they do a good job of writing secure code using the unsafe APIs. Our experience contrasts with existing research literature, which focuses on reducing the false-positive and false-negative rates
    - (拙訳) 私たちの経験上、フィードバックサイクルが早く、修正が比較的簡単であれば、開発者は安全性が保証されているAPIを進んで受け入れます - もし開発者が安全でないコードを書いていると証明出来ない場合でも、彼らが安全性が保証されていないAPIを使って上手くセキュアコーディングを行えるのだとしても。 この私たちの経験は、過検知と検知漏れ率を下げていくということにフォーカスする、既存の研究とは対照的です。
  - To handle complicated use cases, at Google we allow a way to bypass the type restrictions with approval from a security engineer. For example, our database API has a separate package, unsafequery, that exports a distinct unsafequery.String type, which can be constructed from arbitrary strings and appended to SQL queries. Only a small fraction of our queries use the unchecked APIs.
    - (拙訳) 複雑なユースケースを扱うために、Googleでは、セキュリティエンジニアの承認によって型の制限をバイパスすることを認めています。たとえば、私たちのデータベースAPIは他のパッケージを持っています。unsafequeryパッケージは、任意の文字列から構成してSQLクエリに追記することのできる、unsafequery.String型をエクスポートしています。

- [go-safeweb](https://github.com/google/go-safeweb) のSecurityの考え方も勉強になる
  - Secure-by-default
  - Unsafe Usage is Easy to Review, Track and Restrict
  - Designed for Evolving Security Requirements
  - High Compatibility with Go’s Standard Library and Existing Open-Source Frameworks
  - 
  