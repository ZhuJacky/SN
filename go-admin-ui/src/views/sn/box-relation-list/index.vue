<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">
        <el-form ref="queryForm" :model="queryParams" :inline="true" label-width="120px">
          <el-form-item label="箱号" prop="BoxId">
            <el-input
              v-model="queryParams.BoxId"
              placeholder=""
              clearable
              size="small"
              @keyup.enter.native="handleQuery"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="el-icon-search" size="mini" @click="handleQuery">搜索</el-button>
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
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="addBox"
            >测试扫码功能</el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="boxRelationList" border >
          <el-table-column label="序号" width="120" align="center" prop="BoxRelationId" />
          <el-table-column label="箱号" width="120" align="center" prop="BoxId" />
          <el-table-column label="SN号" align="center" prop="SNCode" />
          <el-table-column label="装箱数量" align="center" prop="BoxSum" />
          <el-table-column label="扫码枪IP" width="120" align="center" prop="ScanSource" />
          <el-table-column label="创建时间" align="center" prop="createdAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.createdAt) }}</span>
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
      </el-card>
    </template>
  </BasicLayout>
</template>

<script>
import { listPost,packBox } from '@/api/sn/box-relation-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'BoxRelationInfoManage',
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
      boxRelationList: [],
      // 弹出层标题
      title: '',
      // 是否显示弹出层
      open: false,
      // 状态数据字典
      statusOptions: [],
      // 查询参数
      queryParams: {
        pageIndex: 1,
        pageSize: 100,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      boxData: {
        snCode: undefined,
        status: 3,
        scanSource: '666'
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
    this.getList()

    window.document.onkeypress = (e) => {
      if (window.event) { // IE
        this.nextCode = e.keyCode
      } else if (e.which) { // Netscape/Firefox/Opera
        this.nextCode = e.which
      }
 
      if (e.which === 13) { // 键盘回车事件
        if (this.code.length < 3) return // 扫码枪的速度很快，手动输入的时间不会让code的长度大于2，所以这里不会对扫码枪有效
        console.log('扫码结束。')
        console.log('条形码：', this.code)
        this.parseQRCode(this.code) // 获取到扫码枪输入的内容，做别的操作
        this.lastCode = ''
        this.lastTime = ''
        return
      }
 
      this.nextTime = new Date().getTime()
      if (!this.lastTime && !this.lastCode) {
        this.code = '' // 清空上次的条形码
        this.code += e.key
        console.log('扫码开始---', this.code)
      }
      if (this.lastCode && this.lastTime && this.nextTime - this.lastTime > 500) { // 当扫码前有keypress事件时,防止首字缺失
        this.code = e.key
        console.log('防止首字缺失。。。', this.code)
      } else if (this.lastCode && this.lastTime) {
        this.code += e.key
        console.log('扫码中。。。', this.code)
      }
      this.lastCode = this.nextCode
      this.lastTime = this.nextTime
    }
  },
  methods: {

    /** 查询装箱列表 */
    getList() {
      this.loading = true
      listPost(this.queryParams).then(response => {
        this.boxRelationList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },
    /** 搜索按钮操作 */
    handleQuery() {
      this.queryParams.pageIndex = 1
      this.getList()
    },
    
    /** 当SN码有异常时，触发告警声音 */
    warningAudio() {
      //this.audio = new Audio()
      //this.audio.src  = "http://127.0.0.1:8000/static/audios/do_wrong.mp3"
      //this.audio.play()
    },
    parseQRCode(code) {
      if (code.length === 13) {
        // 处理自己的逻辑
        //alert('条码合法：' + code)
        this.$emit('barCode', code) //通知父组件
         this.addBox(code)
      } else if (code.length === 23) {
        console.log('B类条码:' + code)
      } else if (code.length === 0) {
        console.log('请输入条码！')
      } else {
        this.addBox(code)
        //alert('条码不合法：' + code)
      }
      this.codeValue = code
      // 发送网络请求
    },
    addBox(code) { //开始装箱
      
      //alert('开始装箱：' + code)

      code = '8930500A000024'
      //alert(this.boxData.snCode)
  
      packBox(this.boxData, code).then(response => {
          if (response.code === 200) {              
              let Status = response.data.Status
              if(Status==1) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==2) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==3) {
                this.warningAudio()
                this.msgError(response.msg)
              } else if(Status==0) { //装箱成功
                this.queryParams.BoxId = response.data.BoxId;//填充查询条件
                this.getList()
              } else if(Status==4) { //装满一箱，调用打印机执行打印动作
                this.queryParams.BoxId = response.data.BoxId;//填充查询条件
                this.getList()
              }

          } else {
              this.msgError(response.msg)
          }
      })
    }
  }
}
</script>
