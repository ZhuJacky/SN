<template>
  <BasicLayout>
    <template #wrapper>
      <el-card class="box-card">

        <el-row :gutter="10" class="mb8">
          <el-col :span="1.5">
            <el-button
              v-permisaction="['admin:sysPost:export']"
              type="warning"
              icon="el-icon-microphone"
              size="mini"
              @click="doEnWarehouse"
            >测试入库功能</el-button>
          </el-col>
        </el-row>

        <el-table v-loading="loading" :data="boxList" border >
          <el-table-column label="箱号" width="60" align="center" prop="BoxId" />
          <el-table-column label="批号(LOT)" width="140" align="center" prop="BatchCode" />
          <el-table-column label="产品型号" width="80" align="center" prop="ProductCode" />
          <el-table-column label="UDI号" align="center" prop="UDI" />
          <el-table-column label="工单号" width="150" align="center" prop="WorkCode" />
          <el-table-column label="装箱数量" width="150" align="center" prop="BoxSum" />
          <el-table-column label="状态" width="80" align="center" prop="Status" :formatter="statusFormat">
            <template slot-scope="scope">
              <el-tag
                :type="scope.row.Status === 0 ? 'danger' : 'success'"
                disable-transitions
              >{{ statusFormat(scope.row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="入库时间" align="center" prop="updatedAt" width="155">
            <template slot-scope="scope">
              <span>{{ parseTime(scope.row.updatedAt) }}</span>
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
import { listEnWarehouseBox,doExWarehouseBox } from '@/api/sn/box-info'
import { formatJson } from '@/utils'
import moment from 'moment'

export default {
  name: 'EnWarehouseManage',
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
        pageSize: 20,
        postCode: undefined,
        postName: undefined,
        status: undefined
      },
      boxData: {
        BoxId: undefined,
        Status: 1
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
    this.getDicts('box_info_status').then(response => {
      this.statusOptions = response.data
    })

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
      listEnWarehouseBox(this.queryParams).then(response => {
        this.boxList = response.data.list
        this.total = response.data.count
        this.loading = false
      })
    },

    /** 当SN码有异常时，触发告警声音 001 */
    warningAudio() {
      this.audio = new Audio()
      this.audio.src  = "http://159.75.213.231:8000/static/audios/do_wrong.mp3"
      this.audio.play()
    },

    // 箱子状态翻译
    statusFormat(row) {
      return this.selectDictLabel(this.statusOptions, row.Status)
    }, 
    
    parseQRCode(code) {
      if (code.length === 13) {
        // 处理自己的逻辑
        //alert('条码合法：' + code)
        this.$emit('barCode', code) //通知父组件
         this.doEnWarehouse(code)
      } else if (code.length === 23) {
        console.log('B类条码:' + code)
      } else if (code.length === 0) {
        console.log('请输入条码！')
      } else {
        this.doEnWarehouse(code)
        //alert('条码不合法：' + code)
      }
    },
    
    doEnWarehouse(code) { //执行出库

      //code = '10002'
      this.boxData.BoxId = parseInt(code)

      //alert(code)
      //return

      doExWarehouseBox(this.boxData, code).then(response => {
          if (response.code === 200) {
              let Status = response.data.Status
              if(Status==-1) {
                this.warningAudio()
                this.msgError(response.msg)
              } else {
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
