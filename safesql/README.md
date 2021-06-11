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

#### SQL Injectionを防ぐ静的解析系ツールとの比較 (うろおぼえ!)
- stripe/safesql
  - 基本的にはdb.queryに入る文字列の連結を防ぐアプローチ
  - safesqlはSSAを用いてdb.queryに入れる文字列を検知して、いい感じだが、別の関数で連結された文字列を、db.queryにわたす場合は検知できない

- gosec
  - gosecは誤検知が多いという問題（SQLっぽい文字列連結の検知をする. golangci-lintだとデフォルト無効化されてた気がする）

#### 静的解析での応用(Google)
- Building Secure and Reliable Systemsで紹介されてる
- https://github.com/google/go-safeweb にも説明あり

- アプローチ
  - 基本はSafeSQLを経由してのDB呼び出しを強制する (Secure By Defaultのアプローチで、基本が安全な状態にする)
    - 静的解析によって、safesql使ってなかったらreportingしてSecurity TeamのReviewをいれたり、強制的に使わせたりするアプローチがとれる
    - 危ないことをさせるときは、危ないことが分かるようにする
      - GoのUnsafeも同じような話

- Bulding Secure and Reliable Systems (Google SRE本 No.3)の引用
  - In our experience, as long as the feedback cycle is quick, and fixing each pattern is relatively easy, developers embrace inherently safe APIs much more readily—even when we can’t prove that their code was insecure, or when they do a good job of writing secure code using the unsafe APIs. Our experience contrasts with existing research literature, which focuses on reducing the false-positive and false-negative rates
  - To handle complicated use cases, at Google we allow a way to bypass the type restrictions with approval from a security engineer. For example, our database API has a separate package, unsafequery, that exports a distinct unsafequery.String type, which can be constructed from arbitrary strings and appended to SQL queries. Only a small fraction of our queries use the unchecked APIs.

- [go-safeweb](https://github.com/google/go-safeweb) のSecurityの考え方も勉強になる
  - Secure-by-default
  - Unsafe Usage is Easy to Review, Track and Restrict
  - Designed for Evolving Security Requirements
  - High Compatibility with Go’s Standard Library and Existing Open-Source Frameworks
  - 
  