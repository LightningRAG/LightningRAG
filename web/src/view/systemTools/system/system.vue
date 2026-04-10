<template>
  <div class="system">
    <el-form ref="form" :model="config" label-width="240px">
      <!--  System start  -->
      <el-tabs v-model="activeNames">
        <el-tab-pane :label="$t('tools.systemPage.tab.system')" name="1" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.portVal')">
            <el-input-number
              v-model="config.system.addr"
              :placeholder="$t('tools.systemPage.ph.port')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.dbType')">
            <el-select v-model="config.system['db-type']" class="w-full">
              <el-option value="mysql" />
              <el-option value="pgsql" />
              <el-option value="mssql" />
              <el-option value="sqlite" />
              <el-option value="oracle" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.ossType')">
            <el-select v-model="config.system['oss-type']" class="w-full">
              <el-option value="local" :label="$t('tools.systemPage.ossOpt.local')" />
              <el-option value="qiniu" :label="$t('tools.systemPage.ossOpt.qiniu')" />
              <el-option value="tencent-cos" :label="$t('tools.systemPage.ossOpt.tencentCos')" />
              <el-option value="aliyun-oss" :label="$t('tools.systemPage.ossOpt.aliyunOss')" />
              <el-option value="huawei-obs" :label="$t('tools.systemPage.ossOpt.huaweiObs')" />
              <el-option value="cloudflare-r2" :label="$t('tools.systemPage.ossOpt.cloudflareR2')" />
              <el-option value="minio">MinIO</el-option>
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.multiLogin')">
            <el-switch v-model="config.system['use-multipoint']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.useRedis')">
            <el-switch v-model="config.system['use-redis']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.useMongo')">
            <el-switch v-model="config.system['use-mongo']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.strictAuth')">
            <el-switch v-model="config.system['use-strict-auth']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.rateLimitCount')">
            <el-input-number v-model.number="config.system['iplimit-count']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.rateLimitTime')">
            <el-input-number v-model.number="config.system['iplimit-time']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.disableAutoMigrate')">
            <el-switch v-model="config.system['disable-auto-migrate']" />
          </el-form-item>
          <el-tooltip
            :content="$t('tools.systemPage.routerPrefixTip')"
            placement="top-start"
          >
            <el-form-item :label="$t('tools.systemPage.f.routerPrefix')">
              <el-input
                v-model.trim="config.system['router-prefix']"
                :placeholder="$t('tools.systemPage.ph.routerPrefix')"
              />
            </el-form-item>
          </el-tooltip>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.jwt')" name="2" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.tab.jwt')">
            <el-input
              v-model.trim="config.jwt['signing-key']"
              :placeholder="$t('tools.systemPage.ph.jwtSigning')"
            >
              <template #append>
                <el-button @click="getUUID">{{ $t('tools.systemPage.btnGen') }}</el-button>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.expires')">
            <el-input
              v-model.trim="config.jwt['expires-time']"
              :placeholder="$t('tools.systemPage.ph.expires')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.buffer')">
            <el-input
              v-model.trim="config.jwt['buffer-time']"
              :placeholder="$t('tools.systemPage.ph.buffer')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.issuer')">
            <el-input
              v-model.trim="config.jwt.issuer"
              :placeholder="$t('tools.systemPage.ph.issuer')"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.zap')" name="3" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.level')">
            <el-select v-model="config.zap.level">
              <el-option value="off" :label="$t('tools.systemPage.log.off')" />
              <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
              <el-option value="error" :label="$t('tools.systemPage.log.error')" />
              <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
              <el-option value="info" :label="$t('tools.systemPage.log.info')" />
              <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
              <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.output')">
            <el-select v-model="config.zap.format">
              <el-option value="console" label="console" />
              <el-option value="json" label="json" />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.logPrefix')">
            <el-input
              v-model.trim="config.zap.prefix"
              :placeholder="$t('tools.systemPage.ph.logPrefix')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.logDir')">
            <el-input
              v-model.trim="config.zap.director"
              :placeholder="$t('tools.systemPage.ph.logDir')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.encodeLevel')">
            <el-select v-model="config.zap['encode-level']" class="w-6/12">
              <el-option
                value="LowercaseLevelEncoder"
                label="LowercaseLevelEncoder"
              />
              <el-option
                value="LowercaseColorLevelEncoder"
                label="LowercaseColorLevelEncoder"
              />
              <el-option
                value="CapitalLevelEncoder"
                label="CapitalLevelEncoder"
              />
              <el-option
                value="CapitalColorLevelEncoder"
                label="CapitalColorLevelEncoder"
              />
            </el-select>
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.stackName')">
            <el-input
              v-model.trim="config.zap['stacktrace-key']"
              :placeholder="$t('tools.systemPage.ph.stackName')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.retentionDay')">
            <el-input-number v-model="config.zap['retention-day']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.showLine')">
            <el-switch v-model="config.zap['show-line']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.logToConsole')">
            <el-switch v-model="config.zap['log-in-console']" />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane
          label="Redis"
          name="4"
          class="mt-3.5"
          v-if="config.system['use-redis']"
        >
          <el-form-item :label="$t('tools.systemPage.f.redisDb')">
            <el-input-number v-model="config.redis.db" min="0" max="16" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.addr')">
            <el-input
              v-model.trim="config.redis.addr"
              :placeholder="$t('tools.systemPage.ph.addr')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.password')">
            <el-input
              v-model.trim="config.redis.password"
              :placeholder="$t('tools.systemPage.ph.password')"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.email')" name="5" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.emailTo')">
            <el-input
              v-model="config.email.to"
              :placeholder="$t('tools.systemPage.ph.emailTo')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.port')">
            <el-input-number v-model="config.email.port" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.emailFrom')">
            <el-input
              v-model.trim="config.email.from"
              :placeholder="$t('tools.systemPage.ph.emailFrom')"
            />
          </el-form-item>
          <el-form-item label="host">
            <el-input
              v-model.trim="config.email.host"
              :placeholder="$t('tools.systemPage.ph.host')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.isSsl')">
            <el-switch v-model="config.email['is-ssl']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.loginAuth')">
            <el-switch v-model="config.email['is-loginauth']" />
          </el-form-item>
          <el-form-item label="secret">
            <el-input
              v-model.trim="config.email.secret"
              :placeholder="$t('tools.systemPage.ph.secret')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.testEmail')">
            <el-button @click="email">{{ $t('tools.systemPage.btnTestEmail') }}</el-button>
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane
          :label="$t('tools.systemPage.tab.mongo')"
          name="14"
          class="mt-3.5"
          v-if="config.system['use-mongo']"
        >
          <el-form-item :label="$t('tools.systemPage.f.mongoColl')">
            <el-input
              v-model.trim="config.mongo.coll"
              :placeholder="$t('tools.systemPage.ph.mongoColl')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.mongoOpts')">
            <el-input
              v-model.trim="config.mongo.options"
              :placeholder="$t('tools.systemPage.ph.mongoOpts')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.mongoDb')">
            <el-input
              v-model.trim="config.mongo.database"
              :placeholder="$t('tools.systemPage.ph.dbName')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.username')">
            <el-input
              v-model.trim="config.mongo.username"
              :placeholder="$t('tools.systemPage.ph.username')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.password')">
            <el-input
              v-model.trim="config.mongo.password"
              :placeholder="$t('tools.systemPage.ph.password')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.minPool')">
            <el-input-number v-model="config.mongo['min-pool-size']" min="0" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.maxPool')">
            <el-input-number
              v-model="config.mongo['max-pool-size']"
              min="100"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.socketTimeout')">
            <el-input-number
              v-model="config.mongo['socket-timeout-ms']"
              min="0"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.connectTimeout')">
            <el-input-number
              v-model="config.mongo['socket-timeout-ms']"
              min="0"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.mongoZapLog')">
            <el-switch v-model="config.mongo['is-zap']" />
          </el-form-item>
          <el-form-item
            v-for="(item, k) in config.mongo.hosts"
            :key="k"
            :label="$t('tools.systemPage.mongoNode', { n: k + 1 })"
          >
            <div v-for="(_, k2) in item" :key="k2">
              <el-form-item :key="k + k2" :label="k2" label-width="60">
                <el-input
                  v-model.trim="item[k2]"
                  :placeholder="k2 === 'host' ? $t('tools.systemPage.ph.addr') : $t('tools.systemPage.ph.inputPort')"
                />
              </el-form-item>
            </div>
            <el-form-item v-if="k > 0">
              <el-button
                type="danger"
                size="small"
                plain
                :icon="Minus"
                @click="removeNode(k)"
                class="ml-3"
              />
            </el-form-item>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              size="small"
              plain
              :icon="Plus"
              @click="addNode"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.captcha')" name="7" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.captchaLen')">
            <el-input-number
              v-model="config.captcha['key-long']"
              :min="4"
              :max="6"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.imgWidth')">
            <el-input-number v-model.number="config.captcha['img-width']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.imgHeight')">
            <el-input-number v-model.number="config.captcha['img-height']" />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.database')" name="9" class="mt-3.5">
          <template v-if="config.system['db-type'] === 'mysql'">
            <el-form-item label="">
              <h3>MySQL</h3>
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.username')">
              <el-input
                v-model.trim="config.mysql.username"
                :placeholder="$t('tools.systemPage.ph.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.password')">
              <el-input
                v-model.trim="config.mysql.password"
                :placeholder="$t('tools.systemPage.ph.password')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.addr')">
              <el-input
                v-model.trim="config.mysql.path"
                :placeholder="$t('tools.systemPage.ph.addr')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.dbName')">
              <el-input
                v-model.trim="config.mysql['db-name']"
                :placeholder="$t('tools.systemPage.ph.mysqlDbName')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.prefix')">
              <el-input
                v-model.trim="config.mysql['prefix']"
                :placeholder="$t('tools.systemPage.ph.prefixEmpty')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.pluralTable')">
              <el-switch v-model="config.mysql['singular']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.engine')">
              <el-input
                v-model.trim="config.mysql['engine']"
                :placeholder="$t('tools.systemPage.ph.engineInnodb')"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number
                v-model="config.mysql['max-idle-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number
                v-model="config.mysql['max-open-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.writeLog')">
              <el-switch v-model="config.mysql['log-zap']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.logMode')">
              <el-select v-model="config.mysql['log-mode']">
                <el-option value="off" :label="$t('tools.systemPage.log.off')" />
                <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
                <el-option value="error" :label="$t('tools.systemPage.log.error')" />
                <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
                <el-option value="info" :label="$t('tools.systemPage.log.info')" />
                <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
                <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'pgsql'">
            <el-form-item label="">
              <h3>PostgreSQL</h3>
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.username')">
              <el-input
                v-model="config.pgsql.username"
                :placeholder="$t('tools.systemPage.ph.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.password')">
              <el-input
                v-model="config.pgsql.password"
                :placeholder="$t('tools.systemPage.ph.password')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.addr')">
              <el-input
                v-model.trim="config.pgsql.path"
                :placeholder="$t('tools.systemPage.ph.addr')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.db')">
              <el-input
                v-model.trim="config.pgsql['db-name']"
                :placeholder="$t('tools.systemPage.ph.databaseConn')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.prefix')">
              <el-input
                v-model.trim="config.pgsql['prefix']"
                :placeholder="$t('tools.systemPage.ph.prefix')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.pluralTable')">
              <el-switch v-model="config.pgsql['singular']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.engine')">
              <el-input
                v-model.trim="config.pgsql['engine']"
                :placeholder="$t('tools.systemPage.ph.engine')"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.pgsql['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.pgsql['max-open-conns']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.writeLog')">
              <el-switch v-model="config.pgsql['log-zap']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.logMode')">
              <el-select v-model="config.pgsql['log-mode']">
                <el-option value="off" :label="$t('tools.systemPage.log.off')" />
                <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
                <el-option value="error" :label="$t('tools.systemPage.log.error')" />
                <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
                <el-option value="info" :label="$t('tools.systemPage.log.info')" />
                <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
                <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'mssql'">
            <el-form-item label="">
              <h3>MsSQL</h3>
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.username')">
              <el-input
                v-model.trim="config.mssql.username"
                :placeholder="$t('tools.systemPage.ph.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.password')">
              <el-input
                v-model.trim="config.mssql.password"
                :placeholder="$t('tools.systemPage.ph.password')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.addr')">
              <el-input
                v-model.trim="config.mssql.path"
                :placeholder="$t('tools.systemPage.ph.addr')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.port')">
              <el-input
                v-model.trim="config.mssql.port"
                :placeholder="$t('tools.systemPage.ph.inputPort')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.db')">
              <el-input
                v-model.trim="config.mssql['db-name']"
                :placeholder="$t('tools.systemPage.ph.databaseConn')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.prefix')">
              <el-input
                v-model.trim="config.mssql['prefix']"
                :placeholder="$t('tools.systemPage.ph.prefix')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.pluralTable')">
              <el-switch v-model="config.mssql['singular']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.engine')">
              <el-input
                v-model.trim="config.mssql['engine']"
                :placeholder="$t('tools.systemPage.ph.engine')"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.mssql['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.mssql['max-open-conns']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.writeLog')">
              <el-switch v-model="config.mssql['log-zap']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.logMode')">
              <el-select v-model="config.mssql['log-mode']">
                <el-option value="off" :label="$t('tools.systemPage.log.off')" />
                <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
                <el-option value="error" :label="$t('tools.systemPage.log.error')" />
                <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
                <el-option value="info" :label="$t('tools.systemPage.log.info')" />
                <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
                <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'sqlite'">
            <el-form-item label="">
              <h3>sqlite</h3>
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.username')">
              <el-input
                v-model.trim="config.sqlite.username"
                :placeholder="$t('tools.systemPage.ph.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.password')">
              <el-input
                v-model.trim="config.sqlite.password"
                :placeholder="$t('tools.systemPage.ph.password')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.addr')">
              <el-input
                v-model.trim="config.sqlite.path"
                :placeholder="$t('tools.systemPage.ph.addr')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.port')">
              <el-input
                v-model.trim="config.sqlite.port"
                :placeholder="$t('tools.systemPage.ph.inputPort')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.db')">
              <el-input
                v-model.trim="config.sqlite['db-name']"
                :placeholder="$t('tools.systemPage.ph.databaseConn')"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number v-model="config.sqlite['max-idle-conns']" />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number v-model="config.sqlite['max-open-conns']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.writeLog')">
              <el-switch v-model="config.sqlite['log-zap']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.logMode')">
              <el-select v-model="config.sqlite['log-mode']">
                <el-option value="off" :label="$t('tools.systemPage.log.off')" />
                <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
                <el-option value="error" :label="$t('tools.systemPage.log.error')" />
                <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
                <el-option value="info" :label="$t('tools.systemPage.log.info')" />
                <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
                <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
              </el-select>
            </el-form-item>
          </template>
          <template v-if="config.system['db-type'] === 'oracle'">
            <el-form-item label="">
              <h3>oracle</h3>
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.username')">
              <el-input
                v-model.trim="config.oracle.username"
                :placeholder="$t('tools.systemPage.ph.username')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.password')">
              <el-input
                v-model.trim="config.oracle.password"
                :placeholder="$t('tools.systemPage.ph.password')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.addr')">
              <el-input
                v-model.trim="config.oracle.path"
                :placeholder="$t('tools.systemPage.ph.addr')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.dbName')">
              <el-input
                v-model.trim="config.oracle['db-name']"
                :placeholder="$t('tools.systemPage.ph.mysqlDbName')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.prefix')">
              <el-input
                v-model.trim="config.oracle['prefix']"
                :placeholder="$t('tools.systemPage.ph.prefixEmpty')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.pluralTable')">
              <el-switch v-model="config.oracle['singular']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.engine')">
              <el-input
                v-model.trim="config.oracle['engine']"
                :placeholder="$t('tools.systemPage.ph.engineInnodb')"
              />
            </el-form-item>
            <el-form-item label="maxIdleConns">
              <el-input-number
                v-model="config.oracle['max-idle-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item label="maxOpenConns">
              <el-input-number
                v-model="config.oracle['max-open-conns']"
                :min="1"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.writeLog')">
              <el-switch v-model="config.oracle['log-zap']" />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.logMode')">
              <el-select v-model="config.oracle['log-mode']">
                <el-option value="off" :label="$t('tools.systemPage.log.off')" />
                <el-option value="fatal" :label="$t('tools.systemPage.log.fatal')" />
                <el-option value="error" :label="$t('tools.systemPage.log.error')" />
                <el-option value="warn" :label="$t('tools.systemPage.log.warn')" />
                <el-option value="info" :label="$t('tools.systemPage.log.info')" />
                <el-option value="debug" :label="$t('tools.systemPage.log.debug')" />
                <el-option value="trace" :label="$t('tools.systemPage.log.trace')" />
              </el-select>
            </el-form-item>
          </template>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.oss')" name="10" class="mt-3.5">
          <template v-if="config.system['oss-type'] === 'local'">
            <h2>{{ $t('tools.systemPage.h2.local') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.localPublic')">
              <el-input
                v-model.trim="config.local.path"
                :placeholder="$t('tools.systemPage.ph.localPublic')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.localStore')">
              <el-input
                v-model.trim="config.local['store-path']"
                :placeholder="$t('tools.systemPage.ph.localStore')"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'qiniu'">
            <h2>{{ $t('tools.systemPage.h2.qiniu') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.zone')">
              <el-input
                v-model.trim="config.qiniu.zone"
                :placeholder="$t('tools.systemPage.ph.zone')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.bucketName')">
              <el-input
                v-model.trim="config.qiniu.bucket"
                :placeholder="$t('tools.systemPage.ph.bucket')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.cdnDomain')">
              <el-input
                v-model.trim="config.qiniu['img-path']"
                :placeholder="$t('tools.systemPage.ph.cdn')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.useHttps')">
              <el-switch v-model="config.qiniu['use-https']">{{ $t('tools.systemPage.switchOn') }}</el-switch>
            </el-form-item>
            <el-form-item label="accessKey">
              <el-input
                v-model.trim="config.qiniu['access-key']"
                :placeholder="$t('tools.systemPage.ph.accessKey')"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config.qiniu['secret-key']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.cdnUploadAccel')">
              <el-switch v-model="config.qiniu['use-cdn-domains']" />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'tencent-cos'">
            <h2>{{ $t('tools.systemPage.h2.tencentCos') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.bucket')">
              <el-input
                v-model.trim="config['tencent-cos']['bucket']"
                :placeholder="$t('tools.systemPage.ph.storageBucket')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.region')">
              <el-input
                v-model.trim="config['tencent-cos'].region"
                :placeholder="$t('tools.systemPage.ph.region')"
              />
            </el-form-item>
            <el-form-item label="secretID">
              <el-input
                v-model.trim="config['tencent-cos']['secret-id']"
                :placeholder="$t('tools.systemPage.ph.secretId')"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config['tencent-cos']['secret-key']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.pathPrefix')">
              <el-input
                v-model.trim="config['tencent-cos']['path-prefix']"
                :placeholder="$t('tools.systemPage.ph.pathPrefix')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.visitDomain')">
              <el-input
                v-model.trim="config['tencent-cos']['base-url']"
                :placeholder="$t('tools.systemPage.ph.visitDomain')"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'aliyun-oss'">
            <h2>{{ $t('tools.systemPage.h2.aliyunOss') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.regionShort')">
              <el-input
                v-model.trim="config['aliyun-oss'].endpoint"
                :placeholder="$t('tools.systemPage.ph.endpoint')"
              />
            </el-form-item>
            <el-form-item label="accessKeyId">
              <el-input
                v-model.trim="config['aliyun-oss']['access-key-id']"
                :placeholder="$t('tools.systemPage.ph.accessKeyId')"
              />
            </el-form-item>
            <el-form-item label="accessKeySecret">
              <el-input
                v-model.trim="config['aliyun-oss']['access-key-secret']"
                :placeholder="$t('tools.systemPage.ph.accessKeySecret')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.bucket')">
              <el-input
                v-model.trim="config['aliyun-oss']['bucket-name']"
                :placeholder="$t('tools.systemPage.ph.storageBucket')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.visitDomain')">
              <el-input
                v-model.trim="config['aliyun-oss']['bucket-url']"
                :placeholder="$t('tools.systemPage.ph.visitDomain')"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'huawei-obs'">
            <h2>{{ $t('tools.systemPage.h2.huaweiObs') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.path')">
              <el-input
                v-model.trim="config['hua-wei-obs'].path"
                :placeholder="$t('tools.systemPage.ph.path')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.bucket')">
              <el-input
                v-model.trim="config['hua-wei-obs'].bucket"
                :placeholder="$t('tools.systemPage.ph.storageBucket')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.regionShort')">
              <el-input
                v-model.trim="config['hua-wei-obs'].endpoint"
                :placeholder="$t('tools.systemPage.ph.endpoint')"
              />
            </el-form-item>
            <el-form-item label="accessKey">
              <el-input
                v-model.trim="config['hua-wei-obs']['access-key']"
                :placeholder="$t('tools.systemPage.ph.accessKey')"
              />
            </el-form-item>
            <el-form-item label="secretKey">
              <el-input
                v-model.trim="config['hua-wei-obs']['secret-key']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'cloudflare-r2'">
            <h2>{{ $t('tools.systemPage.h2.r2') }}</h2>
            <el-form-item :label="$t('tools.systemPage.f.path')">
              <el-input
                v-model.trim="config['cloudflare-r2'].path"
                :placeholder="$t('tools.systemPage.ph.path')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.bucket')">
              <el-input
                v-model.trim="config['cloudflare-r2'].bucket"
                :placeholder="$t('tools.systemPage.ph.storageBucket')"
              />
            </el-form-item>
            <el-form-item label="Base URL">
              <el-input
                v-model.trim="config['cloudflare-r2']['base-url']"
                :placeholder="$t('tools.systemPage.ph.baseUrl')"
              />
            </el-form-item>
            <el-form-item label="Account ID">
              <el-input
                v-model.trim="config['cloudflare-r2']['account-id']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
            <el-form-item label="Access Key ID">
              <el-input
                v-model.trim="config['cloudflare-r2']['access-key-id']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
            <el-form-item label="Secret Access Key">
              <el-input
                v-model.trim="config['cloudflare-r2']['secret-access-key']"
                :placeholder="$t('tools.systemPage.ph.secretKey')"
              />
            </el-form-item>
          </template>
          <template v-if="config.system['oss-type'] === 'minio'">
            <h2>{{ $t('tools.systemPage.h2.minio') }}</h2>
            <el-form-item label="Endpoint">
              <el-input
                v-model.trim="config.minio.endpoint"
                :placeholder="$t('tools.systemPage.ph.minioEndpoint')"
              />
            </el-form-item>
            <el-form-item label="Access Key ID">
              <el-input
                v-model.trim="config.minio['access-key-id']"
                :placeholder="$t('tools.systemPage.ph.minioAk')"
              />
            </el-form-item>
            <el-form-item label="Access Key Secret">
              <el-input
                v-model.trim="config.minio['access-key-secret']"
                :placeholder="$t('tools.systemPage.ph.minioSk')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.bucket')">
              <el-input
                v-model.trim="config.minio['bucket-name']"
                :placeholder="$t('tools.systemPage.ph.storageBucket')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.visitDomain')">
              <el-input
                v-model.trim="config.minio['bucket-url']"
                :placeholder="$t('tools.systemPage.ph.visitDomain')"
              />
            </el-form-item>
            <el-form-item label="Base Path">
              <el-input
                v-model.trim="config.minio['base-path']"
                :placeholder="$t('tools.systemPage.ph.basePath')"
              />
            </el-form-item>
            <el-form-item :label="$t('tools.systemPage.f.enableSsl')">
              <el-switch v-model="config.minio['use-ssl']" />
            </el-form-item>
          </template>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.excel')" name="11" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.excelTarget')">
            <el-input
              v-model.trim="config.excel.dir"
              :placeholder="$t('tools.systemPage.ph.excelDir')"
            />
          </el-form-item>
        </el-tab-pane>
        <el-tab-pane :label="$t('tools.systemPage.tab.autocode')" name="12" class="mt-3.5">
          <el-form-item :label="$t('tools.systemPage.f.autocodeRestart')">
            <el-switch v-model="config.autocode['transfer-restart']" />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeRoot')">
            <el-input v-model="config.autocode.root" disabled />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeServer')">
            <el-input
              v-model.trim="config.autocode['server']"
              :placeholder="$t('tools.systemPage.ph.server')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSApi')">
            <el-input
              v-model.trim="config.autocode['server-api']"
              :placeholder="$t('tools.systemPage.ph.sapi')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSInit')">
            <el-input
              v-model.trim="config.autocode['server-initialize']"
              :placeholder="$t('tools.systemPage.ph.sinit')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSModel')">
            <el-input
              v-model.trim="config.autocode['server-model']"
              :placeholder="$t('tools.systemPage.ph.smodel')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSReq')">
            <el-input
              v-model.trim="config.autocode['server-request']"
              :placeholder="$t('tools.systemPage.ph.sreq')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSRouter')">
            <el-input
              v-model.trim="config.autocode['server-router']"
              :placeholder="$t('tools.systemPage.ph.srouter')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeSSvc')">
            <el-input
              v-model.trim="config.autocode['server-service']"
              :placeholder="$t('tools.systemPage.ph.ssvc')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeWeb')">
            <el-input
              v-model.trim="config.autocode.web"
              :placeholder="$t('tools.systemPage.ph.web')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeWApi')">
            <el-input
              v-model.trim="config.autocode['web-api']"
              :placeholder="$t('tools.systemPage.ph.wapi')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeWForm')">
            <el-input
              v-model.trim="config.autocode['web-form']"
              :placeholder="$t('tools.systemPage.ph.wform')"
            />
          </el-form-item>
          <el-form-item :label="$t('tools.systemPage.f.autocodeWTable')">
            <el-input
              v-model.trim="config.autocode['web-table']"
              :placeholder="$t('tools.systemPage.ph.wtable')"
            />
          </el-form-item>
        </el-tab-pane>
      </el-tabs>
    </el-form>
    <div class="mt-4">
      <el-button type="primary" @click="update">{{ $t('tools.systemPage.btnUpdate') }}</el-button>
      <el-button type="primary" @click="reload">{{ $t('tools.systemPage.btnReload') }}</el-button>
    </div>
  </div>
</template>

<script setup>
  import { getSystemConfig, reloadSystem, setSystemConfig } from '@/api/system'
  import { ref } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { Minus, Plus } from '@element-plus/icons-vue'
  import { emailTest } from '@/api/email'
  import { CreateUUID } from '@/utils/format'

  defineOptions({
    name: 'Config'
  })

  const { t } = useI18n()
  const activeNames = ref('1')
  const config = ref({
    system: {
      'iplimit-count': 0,
      'iplimit-time': 0
    },
    jwt: {},
    mysql: {},
    mssql: {},
    sqlite: {},
    pgsql: {},
    oracle: {},
    excel: {},
    autocode: {},
    redis: {},
    mongo: {
      coll: '',
      options: '',
      database: '',
      username: '',
      password: '',
      'min-pool-size': '',
      'max-pool-size': '',
      'socket-timeout-ms': '',
      'connect-timeout-ms': '',
      'is-zap': false,
      hosts: [
        {
          host: '',
          port: ''
        }
      ]
    },
    qiniu: {},
    'tencent-cos': {},
    'aliyun-oss': {},
    'hua-wei-obs': {},
    'cloudflare-r2': {},
    minio: {},
    captcha: {},
    zap: {},
    local: {},
    email: {},
    timer: {
      detail: {}
    }
  })

  const initForm = async () => {
    const res = await getSystemConfig()
    if (res.code === 0) {
      config.value = res.data.config
    }
  }
  initForm()
  const reload = () => {
    ElMessageBox.confirm(t('tools.systemPage.reloadConfirm'), t('tools.systemPage.reloadTitle'), {
      confirmButtonText: t('common.ok'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
      .then(async () => {
        const res = await reloadSystem()
        if (res.code === 0) {
          ElMessage({
            type: 'success',
            message: t('tools.systemPage.opOk')
          })
        }
      })
      .catch(() => {
        ElMessage({
          type: 'info',
          message: t('tools.systemPage.reloadCancel')
        })
      })
  }

  const update = async () => {
    const res = await setSystemConfig({ config: config.value })
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: t('tools.systemPage.configSaved')
      })
      await initForm()
    }
  }

  const email = async () => {
    const res = await emailTest()
    if (res.code === 0) {
      ElMessage({
        type: 'success',
        message: t('tools.systemPage.emailOk')
      })
      await initForm()
    } else {
      ElMessage({
        type: 'error',
        message: t('tools.systemPage.emailFail')
      })
    }
  }

  const getUUID = () => {
    config.value.jwt['signing-key'] = CreateUUID()
  }

  const addNode = () => {
    config.value.mongo.hosts.push({
      host: '',
      port: ''
    })
  }

  const removeNode = (index) => {
    config.value.mongo.hosts.splice(index, 1)
  }
</script>

<style lang="scss" scoped>
  .system {
    @apply bg-white p-9 rounded dark:bg-slate-900;
    h2 {
      @apply p-2.5 my-2.5 text-lg shadow;
    }
  }
</style>
