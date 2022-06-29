// Package brand TODO
package brand

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"database/sql"
	"encoding/asn1"
	"encoding/json"
	"fmt"
	"math"
	"mysslee_qcloud/brand/model"
	"mysslee_qcloud/utils/certutils"

	_ "github.com/lib/pq" // pq TODO
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

// DetectSymantecBrandContextKey 处理2018 Symantec被DigiCert收购后新证书链的品牌区分，需要去symantec网站查询并缓存
type DetectSymantecBrandContextKey struct{}

// TrustCAInfos TODO
type TrustCAInfos struct {
	CacheCACerts []*cacheCACert `json:"cache_cacerts"`
}

// 被吊销的ca证书是否需要重新放入证书还是当成黑名单？？
type cacheCACert struct {
	Hash     string   `json:"hash"`
	Revoked  bool     `json:"revoked"`
	SignFrom []string `json:"sign_from"`
	IsRoot   bool     `json:"root"`
}

var (
	// BrandCacheVersion TODO
	BrandCacheVersion int
	dbConn            *sql.DB
)

var ocspChan chan *CABrand // 品牌管道

// CABrand 证书品牌信息
type CABrand struct {
	Hash       string // 证书SHA1指纹
	BrandName  string // 品牌名
	BrandGroup string // 品牌组
	CA         string // 所属CA common name
	CertName   string // CA证书名
	CertBytes  []byte // CA证书数据
	SelfSigned bool   // 是否是自签名
	Revoke     bool   // 是否被吊销
	IsRoot     bool   // 是否是根证书
	Pin        string // 证书的公钥的pin码
	Expired    bool   // 证书过期

	AuthorityKeyId string // 颁发者秘钥标识
	SubjectKeyId   string // 自身秘钥标识

	X509 *x509.Certificate // 运行时加速缓存
	SM2  *sm2.Certificate  // 国密SM2证书

	SignFrom []*CABrand // 签发者
}

// BrandCanUseInVerify TODO
func BrandCanUseInVerify(brand *CABrand) bool {
	// 兼容lets encrypt 证书过期
	if brand.Hash == "DAC9024F54D8F6DF94935FB1732638CA6AD77C13" {
		return true
	}
	if brand.Revoke || brand.Expired {
		return false
	} else {
		if certutils.IsExpired(brand.X509) {
			brand.Expired = true
			return false
		}
	}
	return true
}

// Init TODO
func Init(db *sql.DB) {
	dbConn = db

	// 初始化黑名单
	InitBlack()

	// 获取初始的缓存版本
	BrandCacheVersion = GetBrandCacheVersion()

	// 先从本地缓存中证书信息
	certBrandStore = &CertBrandStore{}
	err := GenerateBrandForInit(certBrandStore)
	if err != nil {
		log.Errorf("加载证书品牌错误:%v", err)
	}
}

// UpdateLocalBrandCache 判断是否需要进行证书缓存的更新
func UpdateLocalBrandCache(version int) {
	// 本地的缓存版本比etcd中的版本要大
	if BrandCacheVersion >= version {
		return
	}

	BrandCacheVersion = version
	// 更新本地缓存
	GenerateBrandRelation(GetCertStore())

}

// GetBrandCacheVersion 获取数据库缓存版本
func GetBrandCacheVersion() (version int) {
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	return 0
	// }
	// defer db.Close()

	err := dbConn.QueryRow("SELECT version from config").Scan(&version)
	if err != nil {
		return 0
	}
	return
}

// GenerateBrandForInit 从本地json文件加载CA品牌信息
func GenerateBrandForInit(certBrandStore *CertBrandStore) error {
	brands, auths, err := loadCABrand()

	if len(brands) == 0 {
		log.Panic("没有获取任何的证书信息")
	}
	if err != nil {
		return err
	}

	// 从缓存文件中获取证书链信息
	fileCacheLen, cacheTrustRoots, cacheTrustCAs, err := ReadRelationFormLocalCache(brands)
	if err != nil {
		log.Panicf("从缓存中恢复证书链信息错误:%v", err)
	}

	certBrandStore.subjectKeyIdToCABrands = auths
	certBrandStore.caBrands = brands

	dbCacheLen, trustRoots, trustCAs, err := ReadRelationFromDb(brands)
	if err != nil {
		certBrandStore.trustRoots = cacheTrustRoots
		certBrandStore.trustCAs = cacheTrustCAs
		certBrandStore.relationCacheLen = fileCacheLen
		return nil
	}

	// 如果证书链长度变换超过了原始信息的1/2，则使用原来的信息
	if math.Abs(float64(fileCacheLen-dbCacheLen)) > float64(fileCacheLen/2) {
		certBrandStore.trustRoots = cacheTrustRoots
		certBrandStore.trustCAs = cacheTrustCAs
		certBrandStore.relationCacheLen = fileCacheLen
	} else {
		certBrandStore.trustRoots = trustRoots
		certBrandStore.trustCAs = trustCAs
		certBrandStore.relationCacheLen = dbCacheLen
	}

	return nil
}

// GenerateBrandRelation TODO
// 从pg中加载证书缓存和链关系
func GenerateBrandRelation(certBrandStore *CertBrandStore) (err error) {
	brands, auths, err := loadCABrand()
	if err != nil {
		return errors.Wrap(err, "从数据库中加载证书品牌")
	}

	updateLen, trustRoots, trustCAs, err := ReadRelationFromDb(brands)
	if err != nil {
		return errors.Wrap(err, "构建证书链缓存")
	}
	// 如果变化过大，不更新证书缓存
	if math.Abs(float64(certBrandStore.relationCacheLen-updateLen)) > float64(certBrandStore.relationCacheLen/2) {
		return errors.New("缓存文件变换过大")
	}

	certBrandStore.subjectKeyIdToCABrands = auths
	certBrandStore.caBrands = brands
	certBrandStore.trustRoots = trustRoots
	certBrandStore.trustCAs = trustCAs
	certBrandStore.relationCacheLen = updateLen
	return nil
}

// loadCABrand 加载证书品牌
func loadCABrand() (brands map[string]*CABrand, auths map[string][]*CABrand, err error) {
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	return
	// }
	// defer db.Close()

	rows, err := dbConn.Query("SELECT hash,brand_group,brand_name,cert_name,cert_bytes from ca_brand WHERE disabled<>TRUE")
	if err != nil {
		return
	}
	defer rows.Close()

	brands, auths = make(map[string]*CABrand, 0), make(map[string][]*CABrand, 0)
	for rows.Next() {
		b := &CABrand{}
		err = rows.Scan(&b.Hash, &b.BrandGroup, &b.BrandName, &b.CA, &b.CertBytes)
		if err != nil {
			return
		}

		// 顺带解析其他可用数据
		b.X509, err = x509.ParseCertificate(b.CertBytes)
		if err != nil {
			// TODO: 证书8C6C7A20B48EF3BCB0FCB203008773846611486A-CyberTrust-ABB Intermediate CA 3 解析'名称限制'关键扩展 出错x509: unhandled critical extension
			log.Warnf("load brand parse cert(%v): %v", b.Hash, err)
			continue
		}
		b.CertName = certutils.CertOUString(&b.X509.Subject)
		b.Pin = certutils.GenPin(b.X509.RawSubjectPublicKeyInfo)
		b.AuthorityKeyId = fmt.Sprintf("%X", b.X509.AuthorityKeyId)
		b.SubjectKeyId = fmt.Sprintf("%X", b.X509.SubjectKeyId)

		brands[b.Hash] = b
		if b.SubjectKeyId != "" {
			if sks, ok := auths[b.SubjectKeyId]; !ok {
				auths[b.SubjectKeyId] = []*CABrand{b}
			} else {
				auths[b.SubjectKeyId] = append(sks, b)
			}
		}
	}
	return
}

// getPgCABrands 获取pg数据库中所有证书品牌的数据
func getPgCABrands() []*model.Brand {
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	log.Panicf("打开pg数据库失败:%v", err)
	// }
	// defer db.Close()

	rows, err := dbConn.Query(
		`SELECT hash,brand_group,brand_name,cert_name,cert_bytes,confirmed,created_time,disabled,'comment',kind FROM ca_brand`)
	if err != nil {
		log.Panicf("查询ca_brand表中的数据失败:%v", err)
	}
	defer rows.Close()

	var brands []*model.Brand

	for rows.Next() {
		brand := &model.Brand{}
		err = rows.Scan(&brand.Hash, &brand.BrandGroup, &brand.BrandName, &brand.CertName, &brand.CertBytes, &brand.Confirmed,
			&brand.CreateTime, &brand.Disabled, &brand.Comment, &brand.Kind)
		if err != nil {
			continue
		}
		brands = append(brands, brand)
	}
	return brands
}

// divRootsAndIntermediates 区分出根和中间证书
func divRootsAndIntermediates(certs map[string]*CABrand) (trustRoots []*CABrand, trustCAs []*CABrand, err error) {
	trustRoots = make([]*CABrand, 0)
	trustCAs = make([]*CABrand, 0)

	for _, cert := range certs {
		// 剔除过期证书
		if !BrandCanUseInVerify(cert) {
			continue
		}

		if certutils.CheckRoot(cert.X509) { // 验证是否是根证书
			cert.IsRoot = true
			trustRoots = append(trustRoots, cert)
		} else {
			trustCAs = append(trustCAs, cert)
		}
	}
	return trustRoots, trustCAs, nil
}

// checkIntermediatesCheckChain 查询中间证书的证书链
func checkIntermediatesCheckChain(trustRoots []*CABrand, trustCAs []*CABrand) {

	// 第一步，先直接找到可信根(确保最短链到根)
	for _, intermediate := range trustCAs {
		if intermediate.Revoke { // 如果中间证书以及被吊销，不再继续
			continue
		}
		for _, root := range trustRoots {
			if root.Revoke { // 如果根证书已吊销，不参与组链
				continue
			}
			if err := intermediate.X509.CheckSignatureFrom(root.X509); err == nil {
				intermediate.SignFrom = append(intermediate.SignFrom, root)
			} else {
				switch err.(type) {
				case x509.UnknownAuthorityError:
					log.Debugf("err:%v,intermediate hash:%v, root hash:%v\n", err.Error(), intermediate.Hash, root.Hash)
				case x509.InsecureAlgorithmError:
					log.Debugf("err:%v,intermediate hash:%v, root hash:%v\n", err.Error(), intermediate.Hash, root.Hash)
				case asn1.StructuralError:
					log.Debugf("err:%v,intermediate hash:%v, root hash:%v\n", err.Error(), intermediate.Hash, root.Hash)
				case asn1.SyntaxError:
					log.Debugf("err:%v,intermediate hash:%v, root hash:%v\n", err.Error(), intermediate.Hash, root.Hash)
				default:
					// 验签没有通过
				}
			}
		}
	}

	// 组中间链
	for _, intermediate := range trustCAs {
		if intermediate.Revoke {
			continue
		}
		for _, intermediateca := range trustCAs {
			if intermediateca.Revoke {
				continue
			}
			if err := intermediate.X509.CheckSignatureFrom(intermediateca.X509); err == nil {
				intermediate.SignFrom = append(intermediate.SignFrom, intermediateca)
			} else {
				switch err.(type) {
				case x509.UnknownAuthorityError:
					log.Debugf("err:%v,intermediate hash:%v, intermediateca hash:%v\n", err.Error(), intermediate.Hash,
						intermediateca.Hash)
				case x509.InsecureAlgorithmError:
					log.Debugf("err:%v,intermediate hash:%v, intermediateca hash:%v\n", err.Error(), intermediate.Hash,
						intermediateca.Hash)
				case asn1.StructuralError:
					log.Debugf("err:%v,intermediate hash:%v, intermediateca hash:%v\n", err.Error(), intermediate.Hash,
						intermediateca.Hash)
				case asn1.SyntaxError:
					log.Debugf("err:%v,intermediate hash:%v, intermediateca hash:%v\n", err.Error(), intermediate.Hash,
						intermediateca.Hash)
				default:
					//
				}
			}
		}
	}

}

// ReadRelationFormLocalCache TODO
// 从本地缓存中加载关系
func ReadRelationFormLocalCache(brands map[string]*CABrand) (relationLen int, trustRoots, trustCAs []*CABrand,
	err error) {
	return recoverRelation(brands, []byte(brand_cache))
}

// ReadRelationFromDb 读取缓存，重新构成证书链
func ReadRelationFromDb(brands map[string]*CABrand) (relationLen int, trustRoots, trustCAs []*CABrand, err error) {
	trustRoots, trustCAs = make([]*CABrand, 0), make([]*CABrand, 0)

	// 从config表中读取
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	return
	// }
	// defer db.Close()

	var data string
	err = dbConn.QueryRow("SELECT brands_cache FROM config ").Scan(&data)
	if err != nil {
		return
	}

	return recoverRelation(brands, []byte(data))
}

var testHookDBrealtion *bool
var testHookBrandCache *TrustCAInfos

// recoverRelation 从json数据中恢复证书品牌关系
func recoverRelation(brands map[string]*CABrand, data []byte) (relationLen int, trustRoots, trustCAs []*CABrand,
	err error) {
	var cacheInfos TrustCAInfos
	err = json.Unmarshal(data, &cacheInfos)
	if err != nil {
		return 0, nil, nil, err
	}

	if testHookDBrealtion != nil && *testHookDBrealtion {
		testHookBrandCache = &cacheInfos
	}
	caches := cacheInfos.CacheCACerts
	relationLen = len(caches)
	for _, cache := range caches {
		for _, brand := range brands {
			if brand.Hash == cache.Hash { // 如果是根证书添加根证书中
				if cache.Revoked { // 如果缓存文件中标记证书已经被吊销，同样标记成已吊销
					brand.Revoke = true
				}
				if cache.IsRoot {
					brand.IsRoot = true
					trustRoots = append(trustRoots, brand)
				} else { // 添加到中间证书中
					trustCAs = append(trustCAs, brand)
					for _, signHash := range cache.SignFrom {
						var exist bool
						for _, signFrom := range brand.SignFrom {
							if signFrom.Hash == signHash {
								exist = true
								break
							}
						}
						if !exist {
							brand.SignFrom = append(brand.SignFrom, brands[signHash])
						}
					}
				}
			}
		}
	}
	return
}

// SaveRelationToDB TODO
// 存储证书品牌签发缓存和重组状态
func SaveRelationToDB(relation string, keep, needUpdate bool) error {
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	return err
	// }
	// defer db.Close()

	// 分两种情况
	// 1. 延续
	// 2. 修改
	var id int
	err := dbConn.QueryRow("SELECT id FROM config").Scan(&id)
	switch {
	case err == sql.ErrNoRows: // 不存在插入
		_, err = dbConn.Exec("INSERT INTO config (brands_cache,need_rebuild,version) VALUES ($1,$2,version+1)", relation,
			needUpdate)
	case err != nil:
		return err
	default: // 存在更新
		if keep { // 需要保持
			_, err = dbConn.Exec("UPDATE config SET brands_cache=$1,create_time=now(),version=version+1", relation)
		} else { // 不需要保持
			_, err = dbConn.Exec("UPDATE config SET brands_cache=$1, create_time=now(), need_rebuild=$2,version=version+1",
				relation, needUpdate)
		}
	}
	return err
}

// GenerateRelation TODO
// 生成证书链缓存
func GenerateRelation(trustRoots []*CABrand, trustCAs []*CABrand) (result string, err error) {
	var caches = make([]*cacheCACert, 0)
	for _, root := range trustRoots {
		if !BrandCanUseInVerify(root) { // 证书因吊销，过期等原因不可信，不参与组链
			continue
		}
		cache := &cacheCACert{
			SignFrom: make([]string, 0),
		}
		cache.Hash = root.Hash
		cache.IsRoot = true
		caches = append(caches, cache)
	}

	for _, intermediate := range trustCAs {
		if !BrandCanUseInVerify(intermediate) {
			continue
		}
		cache := &cacheCACert{
			SignFrom: make([]string, 0),
		}
		cache.Hash = intermediate.Hash
		for _, signFrom := range intermediate.SignFrom {
			// 去重
			exist := false
			for _, sf := range cache.SignFrom {
				if sf == signFrom.Hash {
					exist = true
				}
			}
			if !exist {
				cache.SignFrom = append(cache.SignFrom, signFrom.Hash)
			}
		}
		caches = append(caches, cache)
	}

	cacheInfos := TrustCAInfos{}

	cacheInfos.CacheCACerts = caches

	data, err := json.Marshal(cacheInfos)
	if err != nil {
		return "", err
	}

	return string(data), err
}

func sendCABrandInfo() {
	for _, cacert := range certBrandStore.caBrands {
		ocspChan <- cacert // 往管道中发送信息
	}
}

// UpdateCertsWarehouseIntermediate TODO
// 更新证书仓库
func UpdateCertsWarehouseIntermediate(brands []*CABrand) {
	for _, brand := range brands {
		certBrandStore.AddTrustCA(brand)
	}
}

// ConfigNeedUpdate TODO
func ConfigNeedUpdate() bool {
	// pgdb, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	log.Errorf("查询config表时，打开数据库错误:%v", err)
	// 	return false
	// }
	// defer pgdb.Close()

	var needUpdate bool
	err := dbConn.QueryRow("SELECT need_rebuild FROM config").Scan(&needUpdate)
	if err != nil {
		return false
	}
	return needUpdate

}

// GetCertBrand TODO
// 获取证书品牌
func GetCertBrand(ctx context.Context, certs []*x509.Certificate) (brandName, brandGroup string) {
	for i, cert := range certs {
		name, group, find := getCertBrand(cert)
		if find {
			// 修正DigiCert->Symantec品牌
			if i == 0 && name == "DigiCert" {
				if ctx.Value(DetectSymantecBrandContextKey{}) != nil {
					if querySymantecBrand(ctx, cert) {
						name = "Symantec"
					}
				}
			}
			return name, group
		}
	}
	return "Other", "Other"
}

// querySymantecBrand 查询并缓存结果
func querySymantecBrand(ctx context.Context, cert *x509.Certificate) (exists bool) {
	cn := cert.Subject.CommonName
	sn := cert.SerialNumber.Text(16)
	hash := fmt.Sprintf("%X", sha1.Sum(cert.Raw))
	// DB查缓存
	err := dbConn.QueryRow("SELECT is_symantec FROM symantec_brand WHERE hash=$1", hash).Scan(&exists)

	if err == sql.ErrNoRows {
		// 在线查询
		sym := &SymantecBrandQuery{}
		sym.Init(ctx)
		exists = sym.QuerySymantec(cn, sn)

		// 缓存
		dbConn.Exec("INSERT INTO symantec_brand(hash,common_name,is_symantec) VALUES($1,$2,$3)", hash, cn, exists)
	}

	return
}

func getCertBrand(cert *x509.Certificate) (brandName, brandGroup string, find bool) {
	authKeyId := fmt.Sprintf("%X", cert.AuthorityKeyId)
	if cert.IsCA {
		authKeyId = fmt.Sprintf("%X", cert.SubjectKeyId)
		if cabrand, ok := certBrandStore.GetLatestAuthBrand(authKeyId); ok {
			return cabrand.BrandName, cabrand.BrandGroup, true
		} else {
			authKeyId = fmt.Sprintf("%X", cert.AuthorityKeyId)
		}
	}

	if cabrand, ok := certBrandStore.GetLatestAuthBrand(authKeyId); ok {
		return cabrand.BrandName, cabrand.BrandGroup, true
	}
	return
}
