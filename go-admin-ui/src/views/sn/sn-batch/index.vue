<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="批号(LOT)" prop="postCode">
            <el-input
              v-model="queryParams.postCode"
              placeholder="请输入批次编码"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="工单号" prop="workCode">
            <el-input
              v-model="queryParams.workCode"

              placeholder="请输入工单号"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>

          <el-form-item label="状态" prop="status">
            <el-select v-model="queryParams.status" placeholder="批次状态" clearable size="small">
              <el-option
                v-for="dict in statusOptions"
                :key="dict.value"
                :label="dict.label"
                :value="dict.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
            <el-button icon="el-icon-refresh" size="mini" @click="resetQuery">重置</el-button>
          </el-form-item>
        </el-form>

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:add']"
              type="primary"
              icon="el-icon-plus"
              size="mini"
              @click="handleAdd"
            >新增</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:edit']"
              type="success"
              icon="el-icon-edit"
              size="mini"
              :disabled="single"
              @click="handleUpdate"
            >修改</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:remove']"
              type="danger"
              icon="el-icon-delete"
              size="mini"
              :disabled="multiple"
              @click="handleDelete"
            >删除</el-button>
          </el-col>
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-download"
              size="mini"
              @click="handleExport"
            >导出</el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="batchList" border @selection-change="handleSelectionChange">
          <el-table-column type="selection" width="40" align="center" />
          <el-table-column label="序号" width="60" align="center" prop="BatchId" />
          <el-table-column label="批号(LOT)" width="120" align="center" prop="BatchCode" />
          <el-table-column label="数量" width="60" align="center" prop="BatchNumber" />
          <el-table-column label="附加" width="60" align="center" prop="BatchExtra" />
          <el-table-column label="产品型号" width="80" align="center" prop="Product.ProductCode" />
          <el-table-column label="产品名称" align="center" prop="Product.ProductName" />
          <el-table-column label="UDI号" align="center" prop="UDI" />
          <el-table-column label="工单号" align="center" prop="WorkCode" />
          <el-table-column label="SN最小值" align="center" prop="SNMin" />
          <el-table-column label="SN最大值" align="center" prop="SNMax" />
          <el-table-column label="生产月份" align="center" prop="ProductMonth" :formatter="dateFormat">
            <template slot-scope="scope">
              <el-tag>{{ dateFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="BatchImgFile" label="图样pdf" align="center" width="100px">
            <template slot-scope="scope">
            <a :href="scope.row.BatchImgFile" target="_blank">查看</a> 
            <!-- <el-image v-if="scope.row.Product" :src="scope.row.BatchImgFile" :preview-src-list="[scope.row.BatchImgFile]"></el-image> -->
            <!-- <button @click="openpdf">打开pdf</button>-->
            <vueshowpdf  @closepdf="closepdf" v-model="isshowpdf" :pdfurl="scope.row.BatchImgFile"
                @pdferr="pdferr"
                :maxscale='4'
                :minscale='1'
                :scale='1.0'></vueshowpdf>
            <pdf :src="scope.row.BatchImgFile" style="border: 1px solid red"> </pdf>
            </template> 
          </el-table-column>
          <el-table-column label="状态" align="center" prop="status" :formatter="statusFormat">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.status === 1 ? 'danger' : 'success'"
                disable-transitions
              >{{ statusFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="制作类型" align="center" prop="External" :formatter="externalFormat">
          <template slot-scope="scope">
            <el-tag
              :type="scope.row.External === 1 ? 'danger' : 'success'"
              disable-transitions
            >{{ externalFormat(scope.row) }}</el-tag>
          </template>
        </el-table-column>
          <el-table-column label="创建时间" align="center" prop="createdAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" align="center" class-name="small-padding fixed-width">
            <template slot-scope="scope">
             <el-button
               v-permisaction="['admin:sysPost:edit']"
               size="mini"
               type="text"
               icon="el-icon-view"
               @click="handleDetails(scope.row)"
             >SN列表</el-button>
              <el-button
                  v-permisaction="['admin:sysPost:edit']"
                  size="mini"
                  type="text"
                  icon="el-icon-edit"
                  @click="handleUpdate(scope.row)"
                >修改</el-button>
              <el-button
                v-permisaction="['admin:sysPost:remove']"
                size="mini"
                type="text"
                icon="el-icon-delete"
                @click="handleDelete(scope.row)"
              >删除</el-button>
            </template>
          </el-table-column>
        </el-table>

        <pagination
          v-show="total>0"
          :total="total"
          :page.sync="queryParams.pageIndex"
          :limit.sync="queryParams.pageSize"
          @pagination="getList"
        />

        <!-- 添加或修改批次对话框 -->
        <el-dialog :title="title" :visible.sync="open" width="600px">
          <el-form ref="form" :model="form" :rules="rules" label-width="140px">
            <el-form-item  label="格式" prop="snFormat">
              <el-radio-group :disabled="noEdit" v-model="form.snFormat" v-on:input="changeSnFormat()">
                <el-radio
                  :key="0"
                  :label="0"
                >不带括号</el-radio>
                <el-radio
                  :key="1"
                  :label="1"
                >带括号</el-radio>
              </el-radio-group>
              <el-form-item v-if="is_sn_format_show" label="UDI码格式" prop="udiFormatInfo">
                <el-input :disabled="noEdit" v-model="form.udiFormatInfo" placeholder="(01)" />
              </el-form-item>
              <el-form-item v-if="is_sn_format_show" label="LOT号格式" prop="lotFormatInfo">
                <el-input :disabled="noEdit" v-model="form.lotFormatInfo" placeholder="(01)" />
              </el-form-item>
              <el-form-item v-if="is_sn_format_show" label="SN号格式" prop="snFormatInfo">
                <el-input :disabled="noEdit" v-model="form.snFormatInfo" placeholder="(01)" />
              </el-form-item>
            </el-form-item>
            <el-form-item label="批号(LOT)生成方式" prop="batchCodeFormat">
              <el-radio-group :disabled="noEdit" v-model="form.batchCodeFormat" v-on:input="changeBatchCodeFormat()">
                <el-radio
                  :key="0"
                  :label="0"
                >自动生成</el-radio>
                <el-radio
                  :key="1"
                  :label="1"
                >手动填写</el-radio>
              </el-radio-group>
              <el-form-item v-if="is_batch_code_show" label="批号(LOT)" prop="batchCodeFormatInfo">
                <el-input :disabled="noEdit"  v-model="form.batchCodeFormatInfo" placeholder="批号" />
              </el-form-item>
            </el-form-item>
            <el-form-item label="SN生成方式" prop="SNCodeRules">
              <el-radio-group :disabled="noEdit" v-model="form.SNCodeRules" v-on:input="changeSNCodeRulesFormat()">
                <el-radio
                  :key="0"
                  :label="0"
                >自动生成</el-radio>
                <el-radio
                  :key="1"
                  :label="1"
                >客户指定SN号</el-radio>
              </el-radio-group>
              <el-form-item v-if="is_min_sn_code_show" label="最小SN号" prop="minSNCode">
                <el-input :disabled="noEdit"  v-model="form.minSNCode" placeholder="最小SN号" />
              </el-form-item>
              <el-form-item v-if="is_max_sn_code_show" label="最大SN号" prop="MaxSNCode">
              <el-input :disabled="noEdit"  v-model="form.MaxSNCode" placeholder="最大SN号" />
            </el-form-item>
            </el-form-item>
            <el-form-item label="产品名称" prop="ProductName">
              <el-select  :disabled="noEdit" v-model="form.ProductName" placeholder="请选择"  v-on:input="changeForm()">
                  <el-option
                    v-for="dict in productList"
                    :key="dict.ProductName"
                    :label="dict.ProductName"
                    :value="dict.ProductName"
                  />
                </el-select>
            </el-form-item>
            <el-form-item label="产品型号" prop="ProductId">
              <el-select  :disabled="noEdit" v-model="form.ProductId" placeholder="请选择">
                  <el-option
                    v-for="dict in productCodeList"
                    :key="dict.ProductCode"
                    :label="dict.ProductCode"
                    :value="dict.ProductId"
                  />
                </el-select>
            </el-form-item>
            <el-form-item label="UDI号" prop="UDI">
              <el-input :disabled="noEdit" v-model="form.UDI" placeholder="UDI号" />
            </el-form-item>
            <el-form-item label="工单号" prop="WorkCode">
              <el-input :disabled="noEdit" v-model="form.WorkCode" placeholder="工单号" />
            </el-form-item>
            <el-form-item label="制作类型" prop="External">
              <el-radio-group  v-model="form.External">
                <el-radio
                  :key="0"
                  :label="0"
                >自制</el-radio>
                <el-radio
                  :key="1"
                  :label="1"
                >外购</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="批次数量" prop="BatchNumber">
              <el-input-number :disabled="noEdit" readonly="noEdit" v-model="form.BatchNumber" controls-position="right" :min="0" />
            </el-form-item>
            <el-form-item label="多备数量" prop="BatchExtra">
              <el-input-number :disabled="noEdit" readonly="noEdit" v-model="form.BatchExtra" controls-position="right" :min="0" />
            </el-form-item>
            <el-form-item label="批次状态" prop="status">
              <el-select v-model="form.status" placeholder="请选择" >
                  <el-option
                    v-for="dict in statusOptions"
                    :key="dict.value"
                    :label="dict.label"
                    :value="dict.value"
                  />
              </el-select>
            </el-form-item>
            <el-form-item label="图样pdf" prop="ProductSNImage">
            <el-upload class="avatar-uploader" accept=".pdf"
                ref="ProductSNImage" :headers="headers" :file-list="sys_app_logofileList" :action="sys_app_logoAction" style="float:left" :before-upload="sys_app_logoBeforeUpload" list-type="picture-card" :show-file-list="false"  :on-success="uploadSuccess">
              <pdf v-if="form.ProductSNImage" :src="form.ProductSNImage" style="border: 1px solid red; display: inline-block" ></pdf>            
              <i class="el-icon-plus avatar-uploader-icon" ></i>
            </el-upload>
            </el-form-item>
            <el-form-item label="生产月份" prop="ProductMonth">
              <el-date-picker :disabled="noEdit" v-model="form.ProductMonth" type="month" placeholder="选择日期" format="yyyy年MM月" value-format="yyyy-MM" controls-position="right" :min="0" />
            </el-form-item>
            <el-form-item label="备注" prop="remark">
              <el-input v-model="form.Comment" type="textarea" placeholder="请输入内容" />
            </el-form-item>

          </el-form>
          <div slot="footer" class="dialog-footer">
            <el-button type="primary" @click="submitForm">确 定</el-button>
            <el-button @click="cancel">取 消</el-button>
          </div>
        </el-dialog>
      </el-card>
    </template>
  </BasicLayout>
</template>

<script>
import { listPost, getPost, delPost, addPost, updatePost } from '@/api/sn/sn-batch'
import { listProduct } from '@/api/sn/sn-product'
import { formatJson } from '@/utils'
import { getToken } from '@/utils/auth'
import moment from 'moment'
import pdf from 'vue-pdf'
import vueshowpdf from 'vueshowpdf'
function openNewWin(url)
{
	window.open(url);
}
export default {
  components:{vueshowpdf,pdf},
  name: 'SysPostManage',
  data() {
    return {
      headers: { 'Authorization': 'Bearer ' + getToken()},
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      is_batch_code_show:false,
      is_sn_format_show:false,
      is_min_sn_code_show:false,
      is_max_sn_code_show:false,
      isshowpdf: false,
      // 总条数
      total: 0,
      // 岗位表格数据
      batchList: [],
      productList: [],
      productCodeList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      noEdit: false,
      // 状态数据字典
      statusOptions: [],
      // 外协
      externalOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 10,
        postCode: undefined,
        workCode: undefined,
        status: undefined
      },
      sys_app_logoAction: 'http://159.75.213.231/api/v1/public/uploadFile',
      sys_app_logofileList: [],
      // 表单参数
      form: {},

      // 表单校验
      rules: {
        BatchName: [
          { required: true, message: '批次名称不能为空', trigger: 'blur' }
        ],
        ProductId: [
          { required: true, message: '产品名称不能为空', trigger: 'blur' }
        ],
        UDI: [
          { required: true, message: 'UDI不能为空', trigger: 'blur' }
        ],
        minSNCode:[
          { required: true, message: '最小SN号不能为空', trigger: 'blur' }
        ],
        MaxSNCode:[
          { required: true, message: '最大SN号不能为空', trigger: 'blur' }
        ],
        batchCodeFormatInfo:[
          { required: true, message: 'LOT号不能为空', trigger: 'blur' }
        ],
        udiFormatInfo:[
          { required: true, message: 'UDI码不能为空', trigger: 'blur' }
        ],
        lotFormatInfo:[
          { required: true, message: 'LOT号不能为空', trigger: 'blur' }
        ],
        snFormatInfo:[
          { required: true, message: 'SN格式不能为空', trigger: 'blur' }
        ],
        WorkCode: [
          { required: true, message: '工单号不能为空', trigger: 'blur' }
        ],
        BatchNumber: [
          { required: true, message: '批次数量不能为空', trigger: 'blur' }
        ],
        External: [
          { required: true, message: '制作类型不能为空', trigger: 'blur' }
        ]
      }
    }
  },
  created() {
    this.getProductList()
    this.getList()
    this.getDicts('sn_batch_status').then(response => {
      this.statusOptions = response.data
    })
    this.getDicts('sn_batch_external').then(response => {
      this.externalOptions = response.data
    })
  },
  methods: {
    closepdf() {
      this.isshowpdf = false
    },
    pdferr(errurl) {
      console.log(errurl)
    },
    openpdf() {
      this.isshowpdf = true
    },
    uploadSuccess(response, file, fileList) {
      this.$forceUpdate()
      this.form.ProductSNImage = process.env.VUE_APP_BASE_API + response.data.full_path
      this.$forceUpdate()
      console.log(this.form.ProductSNImage)
      console.log(response.data.full_path)
    },
    sys_app_logoBeforeUpload(file) {
      const isRightSize = file.size / 1024 / 1024 < 2
      if (!isRightSize) {
        this.$message.error('文件大小超过 2MB')
      }
      return isRightSize
    },
    /** 查询岗位列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.batchList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },
    getProductList() {
      listProduct().then(response => {
        this.productList = response.data.list
      })
    },
    // 岗位状态字典翻译
    statusFormat(row) {
      return this.selectDictLabel(this.statusOptions, row.status)
    },
    externalFormat(row) {
      return this.selectDictLabel(this.externalOptions, row.External)
    },
    dateFormat(row) {
      return moment(row.ProductMonth).format("YYYY-MM")
    },
    // 取消按钮
    cancel() {
      this.open = false
      this.reset()
    },
    // 表单重置
    reset() {
      this.form = {
        BatchId: undefined,
        postCode: undefined,
        postName: undefined,
        ProductCode: undefined,
        sort: 0,
        status: '1',
        remark: undefined
      }
      this.resetForm('form')
      this.is_batch_code_show=false
      this.is_sn_format_show=false
      this.is_min_sn_code_show=false
      this.is_max_sn_code_show=false

    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageIndex = 1
      this.getList()
    },
    /** 重置按钮操作 */
    resetQuery() {
      this.resetForm('queryForm')
      this.handleQuery()
    },
    // 多选框选中数据
    handleSelectionChange(selection) {
      this.ids = selection.map(item => item.BatchId)
      this.single = selection.length !== 1
      this.multiple = !selection.length
    },
    /** 新增按钮操作 */
    handleAdd() {
      this.reset()
      this.form.snFormat=0
      this.form.batchCodeFormat=0
      this.form.SNCodeRules=0
      this.form.noEdit=false
      this.open = true
      this.title = '添加批次'
      this.noEdit = false
      this.productCodeList = []//重新初始化下拉列表
    },
    /** 修改按钮操作 */
    handleUpdate(row) {
      this.reset()
      const postId = (row.BatchId && [row.BatchId]) || this.ids
      getPost(postId).then(response => {
        this.form = response.data
        this.form.ProductSNImage=row.BatchImgFile
        this.form.status = String(this.form.status)
        this.form.snFormat=row.SNFormat
        this.changeSnFormat()
        this.form.snFormatInfo=row.SNFormatInfo
        this.form.udiFormatInfo=row.UDIFormatInfo
        this.form.lotFormatInfo=row.LOTFormatInfo
        this.changeSNCodeRulesFormat()
        this.form.batchCodeFormat=row.BatchCodeFormat
        this.changeBatchCodeFormat()
        this.form.batchCodeFormatInfo=row.BatchCode
        this.form.minSNCode=row.SNMin
        this.form.MaxSNCode=row.SNMax
        this.form.ProductName=row.Product.ProductName
        this.open = true
        this.title = '修改批次'
        this.noEdit=true
      })
    },
    /*
    changeForm: function() {
      for(let i=0; i<this.productList.length; i++) {
        if(this.productList[i].ProductId === this.form.ProductId) {
          //this.form.UDI=this.productList[i].UDI
          this.form.ProductCode=this.productList[i].ProductCode

        }
      }
    },*/
    changeForm: function() {

      listProduct(this.form.ProductName).then(response => {
        this.productCodeList = response.data.list
        //this.form.ProductId="" //产品重新做了选择，需要重置该表单
      })
    },
    changeSnFormat: function() {
        if(this.form.snFormat===1) {
            this.is_sn_format_show=true
        } else {
            this.is_sn_format_show=false
        }
    },
    changeBatchCodeFormat: function() {
         if(this.form.batchCodeFormat===1) {
             this.is_batch_code_show=true
         } else {
             this.is_batch_code_show=false
         }
     },
     changeSNCodeRulesFormat: function() {
       if(this.form.SNCodeRules===1) {
           this.is_min_sn_code_show=true
           this.is_max_sn_code_show=true
       } else {
           this.is_min_sn_code_show=false
           this.is_max_sn_code_show=false
       }
    },
    /** 提交按钮 */
    submitForm: function() {

      //提交之前，给ProductCode赋值
      for(let i=0; i<this.productCodeList.length; i++) {
        if(this.productCodeList[i].ProductId === this.form.ProductId) {
          this.form.ProductCode=this.productCodeList[i].ProductCode
        }
      }

      this.$refs['form'].validate(valid => {
        if (valid) {
          this.form.status = parseInt(this.form.status)
          if (this.form.BatchId !== undefined) {
            this.form.ProductMonth=this.form.ProductMonth.slice(0,7)
            updatePost(this.form, this.form.BatchId).then(response => {
              if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
              } else {
                this.msgError(response.msg)
              }
            })
          } else {
            addPost(this.form).then(response => {
              if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
              } else {
                this.msgError(response.msg)
              }
            })
          }
        }
      })
    },
    handleDetails(row) {
        //alert(row.Product.ProductCode)
        this.$router.push({path: '/sn/sn-list?batch_code=' + row.BatchCode + "&product_code=" + row.Product.ProductCode});
    },
    /** 删除按钮操作 */
    handleDelete(row) {
      // const postIds = row.postId || this.ids
      const Ids = (row.BatchId && [row.BatchId]) || this.ids
      this.$confirm('是否确认删除批次编号为"' + Ids + '"的数据项?', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(function() {
        return delPost({ 'ids': Ids })
      }).then((response) => {
        if (response.code === 200) {
          this.msgSuccess(response.msg)
          this.open = false
          this.getList()
        } else {
          this.msgError(response.msg)
        }
      }).catch(function() {})
    },
    /** 导出按钮操作 */
    handleExport() {
      // const queryParams = this.queryParams
      this.$confirm('是否确认导出所有批次数据项?', '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.downloadLoading = true
        import('@/vendor/Export2Excel').then(excel => {
          const tHeader = ['批次ID', '批次编码', '批次数量','附加数量','产品型号','UDI号','工单号','SN最大值','SN最小值','生产月份','备注', '状态','创建时间']
          const filterVal = ['BatchId', 'BatchCode', 'BatchNumber','BatchExtra','ProductCode','UDI','WorkCode','SNMax','SNMin', 'ProductMonth','Comment','status','createdAt']
          const list = this.batchList
          const data = formatJson(filterVal, list)
          excel.export_json_to_excel({
            header: tHeader,
            data,
            filename: '批次管理',
            autoWidth: true, // Optional
            bookType: 'xlsx' // Optional
          })
          this.downloadLoading = false
        })
      }).catch(function() {})
    }
  }
}
</script>

<style scoped>
.container {
  font-family: PingFang SC;
  width: 100%;
  height: 500px;
}
</style>

