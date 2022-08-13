<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="批号(LOT)" prop="batchCode">
            <el-input
              v-model="queryParams.batchCode"
              placeholder="请输入批号"
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item label="产品型号" prop="productCode">
            <el-input
              v-model="queryParams.productCode"
              placeholder="请输入产品型号"
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
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="warningAudio"
            ></el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="batchList" border >
          <el-table-column label="序号" width="60" align="center" prop="SNId" />
          <el-table-column label="批号" width="140" align="center" prop="BatchCode" />
          <el-table-column label="产品型号" width="80" align="center" prop="ProductCode" />
          <el-table-column label="UDI号" align="center" prop="UDI" />
          <el-table-column label="工单号" width="150" align="center" prop="WorkCode" />
          <el-table-column label="SN编码" width="200" align="center" prop="SNCode" />
          <el-table-column label="生产月份" width="100" align="center" prop="ProductMonth" :formatter="dateFormat">
            <template slot-scope="scope">
              <el-tag>{{ dateFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>

          <el-table-column label="状态" width="80" align="center" prop="status" :formatter="statusFormat">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.status === 5 ? 'danger' : 'success'"
                disable-transitions
              >{{ statusFormat(scope.row) }}</el-tag>
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
                icon="el-icon-edit"
                @click="handleUpdate(scope.row)"
              >修改</el-button>
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
          <el-form ref="form" :model="form" :rules="rules" label-width="120px">
            <el-form-item label="SN状态" prop="status">
              <el-select v-model="form.status" placeholder="请选择" >
                  <el-option
                    v-for="dict in statusOptions"
                    :key="dict.value"
                    :label="dict.label"
                    :value="dict.value"
                  />
              </el-select>
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
import { listPost, updatePost } from '@/api/sn/sn-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'SNInfoManage',
  data() {
    return {
      // 遮罩层
      loading: true,
      // 选中数组
      ids: [],
      // 非单个禁用
      single: true,
      // 非多个禁用
      multiple: true,
      // 总条数
      total: 0,
      // 岗位表格数据
      batchList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      // 状态数据字典
      statusOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 10,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      // 表单参数
      form: {
      },
      // 表单校验
      rules: {

      }
    }
  },
  created() {

    //填充查询条件
    var query=this.$route.query;
    let batch_code = query.batch_code;
    let product_code = query.product_code;
    this.queryParams.batchCode = batch_code;
    this.queryParams.productCode = product_code;
    this.getList()
    this.getDicts('sn_info_status').then(response => {
      this.statusOptions = response.data
    })
  },
  methods: {
    /** 查询岗位列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.batchList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },
    // SN状态翻译
    statusFormat(row) {
      return this.selectDictLabel(this.statusOptions, row.status)
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
            SNId: undefined,
            status: '1'
          }
          this.resetForm('form')
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
    /** 修改按钮操作 */
    handleUpdate(row) {
        this.reset()
        //const postId = (row.SNId && [row.SNId]) || this.ids
        this.form.SNId = row.SNId
        this.open = true
        this.title = '修改SN'
        //alert(postId)
    },
    handleDetails(row) {
      var query=this.$route.query;
      let batch_code = query.batch_code;
    },

    /** 提交按钮 */
    submitForm: function() {
        this.form.status = parseInt(this.form.status)
        updatePost(this.form, this.form.SNId).then(response => {
            if (response.code === 200) {
                this.msgSuccess(response.msg)
                this.open = false
                this.getList()
            } else {
                this.msgError(response.msg)
            }
        })

    },
    /** 触发告警声音 */
    warningAudio() {
      this.audio = new Audio()
      this.audio.src  = "http://127.0.0.1:8000/static/audios/do_wrong.mp3"
      this.audio.play()
    }
  }
}
</script>
