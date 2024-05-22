package entity

type MasterAsuransi struct {
	NoMsn             string  `form:"no_msn" json:"no_msn"`
	NamaCustomer      string  `form:"nama_customer" json:"nama_customer"`
	Nik               string  `form:"nik" json:"nik"`
	NamaMotor         string  `form:"nama_motor" json:"nama_motor"`
	TglFaktur         string  `form:"tgl_faktur" json:"tgl_faktur"`
	NoTelepon         string  `form:"no_telepon" json:"no_telepon"`
	Status            string  `form:"status" json:"status"`
	AlasanPending     *string `form:"alasan_pending" json:"alasan_pending"`
	StatusBayar       *string `form:"status_bayar" json:"status_bayar"`
	TglBayar          *string `form:"tgl_bayar" json:"tgl_bayar"`
	KdUser            string
	AlasanTdkBerminat *string `form:"alasan_tdk_berminat" json:"alasan_tdk_berminat"`
	KdDlr             *string `form:"kd_dlr" json:"kd_dlr"`
	NmDlr             *string `form:"nm_dlr" json:"nm_dlr"`
	Kelurahan         *string `form:"kelurahan" json:"kelurahan"`
	Kecamatan         *string `form:"kecamatan" json:"kecamatan"`
	Kodepos           *string `form:"kodepos" json:"kodepos"`
	JnsBrg            *string `form:"jns_brg" json:"jns_brg"`
	Harga             int64   `form:"harga" json:"harga"`
	JnsAsuransi       int64   `form:"jenis_asuransi" json:"jenis_asuransi"`
}

type MasterAsuransiReal struct {
	NoMsn             string  `form:"no_msn" json:"no_msn" gorm:"primary_key;column:no_msn"`
	NamaCustomer      string  `form:"nama_customer" json:"nama_customer" gorm:"column:nm_customer11"`
	Nik               string  `form:"nik" json:"nik" gorm:"column:nik"`
	NamaMotor         string  `form:"nama_motor" json:"nama_motor" gorm:"column:nm_mtr"`
	TglFaktur         string  `form:"tgl_faktur" json:"tgl_faktur" gorm:"column:tgl_faktur"`
	NoTelepon         string  `form:"no_telp" json:"no_telp" gorm:"column:no_telp"`
	NoTelepon2        string  `form:"no_telp2" json:"no_telp2" gorm:"column:no_telp2"`
	Status            string  `form:"status" json:"status" gorm:"column:sts_asuransi"`
	AlasanPending     *string `form:"alasan_pending" json:"alasan_pending" gorm:"column:alasan_pending"`
	StatusBayar       *string `form:"status_bayar" json:"status_bayar" gorm:"column:sts_bayar"`
	TglBayar          *string `form:"tgl_bayar" json:"tgl_bayar" gorm:"column:tgl_bayar"`
	AppTransId        string  `form:"app_trans_id" json:"app_trans_id" gorm:"column:app_trans_id"`
	TglLahir          *string `form:"tgl_lahir" json:"tgl_lahir" gorm:"column:tgl_lahir"`
	KdUser            string  `gorm:"column:kd_user"`
	AlasanTdkBerminat *string `form:"alasan_tdk_berminat" json:"alasan_tdk_berminat" gorm:"column:alasan_tdk_berminat"`
	KdDlr             *string `form:"kd_dlr" json:"kd_dlr" gorm:"column:kd_dlr"`
	NmDlr             *string `form:"nm_dlr" json:"nm_dlr" gorm:"column:nm_dlr"`
	Kelurahan         *string `form:"kelurahan" json:"kelurahan" gorm:"column:kelurahan"`
	Kecamatan         *string `form:"kecamatan" json:"kecamatan" gorm:"column:kecamatan"`
	Kodepos           *string `form:"kodepos" json:"kodepos" gorm:"column:kodepos"`
	JnsBrg            string  `form:"jns_brg" json:"jns_brg" gorm:"column:jns_brg"`
	Harga             int64   `form:"harga" json:"harga" gorm:"column:harga"`
	JnsAsuransi       int64   `form:"jenis_asuransi" json:"jenis_asuransi" gorm:"column:jenis_asuransi"`
	JnsSource         string  `form:"jenis_source" json:"jenis_source" gorm:"column:jenis_source"`
	IdTransaksi       string  `form:"id_transaksi" json:"id_transaksi" gorm:"->"`
}

func (MasterAsuransiReal) TableName() string {
	return "asuransi"
}

type MasterAsuransiGorm struct {
	NoMsn string `gorm:"column:no_msn"`
	Nik   string `gorm:"column:nik"`
}

func (MasterAsuransiGorm) TableName() string {
	return "asuransi"
}
