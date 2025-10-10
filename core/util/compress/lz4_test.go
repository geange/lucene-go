package compress

import (
	"os"
	"testing"

	"github.com/geange/lucene-go/core/store"
)

func TestCompress(t *testing.T) {
	ht := NewFastCompressionHashTable()

	buf := store.NewBufferDataOutput()

	t.Log("origin size", len([]byte(data)))

	Compress([]byte(data), buf, ht)

	t.Log(len(buf.Bytes()))

	f, _ := os.Create("lz4.out")
	f.Write(buf.Bytes())
	f.Close()
}

var data = `<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="AutoImportSettings">
    <option name="autoReloadType" value="ALL" />
  </component>
  <component name="ChangeListManager">
    <list default="true" id="6d1f9e8e-b1fa-4efd-93bd-11b313d7dc2f" name="更改" comment="chore: DocValuesConsumer.writeValuesMultipleBlocks" />
    <option name="SHOW_DIALOG" value="false" />
    <option name="HIGHLIGHT_CONFLICTS" value="true" />
    <option name="HIGHLIGHT_NON_ACTIVE_CHANGELIST" value="false" />
    <option name="LAST_RESOLUTION" value="IGNORE" />
  </component>
  <component name="FileTemplateManagerImpl">
    <option name="RECENT_TEMPLATES">
      <list>
        <option value="Go File" />
      </list>
    </option>
  </component>
  <component name="GOROOT" url="file:///usr/local/go" />
  <component name="Git.Settings">
    <option name="RECENT_BRANCH_BY_REPOSITORY">
      <map>
        <entry key="$PROJECT_DIR$" value="main" />
      </map>
    </option>
    <option name="RECENT_GIT_ROOT_PATH" value="$PROJECT_DIR$" />
    <option name="UPDATE_TYPE" value="REBASE" />
  </component>
  <component name="HighlightingSettingsPerFile">
    <setting file="file:///usr/local/go/src/math/bits/bits.go" root0="SKIP_INSPECTION" />
  </component>
  <component name="ProjectColorInfo">{
  &quot;associatedIndex&quot;: 3
}</component>
  <component name="ProjectId" id="30pYhxzBo6o1AcOf2quQwkNgHpS" />
  <component name="ProjectViewState">
    <option name="hideEmptyMiddlePackages" value="true" />
    <option name="showLibraryContents" value="true" />
  </component>
  <component name="PropertiesComponent">{
  &quot;keyToString&quot;: {
    &quot;DefaultGoTemplateProperty&quot;: &quot;Go File&quot;,
    &quot;Go 测试.github.com/geange/lucene-go/core/util/automaton 中的 TestFrozenIntSet_Equals/TC04_-_values_differ.executor&quot;: &quot;Debug&quot;,
    &quot;ModuleVcsDetector.initialDetectionPerformed&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.GoLinterPluginOnboarding&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.GoLinterPluginStorageMigration&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.ShowReadmeOnStart&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.TerminalTabsStorage.copyFrom.TerminalArrangementManager&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.TerminalTabsStorage.copyFrom.TerminalArrangementManager.252&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.git.unshallow&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.go.formatter.settings.were.checked&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.go.migrated.go.modules.settings&quot;: &quot;true&quot;,
    &quot;RunOnceActivity.go.modules.go.list.on.any.changes.was.set&quot;: &quot;true&quot;,
    &quot;git-widget-placeholder&quot;: &quot;feature/lucene80&quot;,
    &quot;go.import.settings.migrated&quot;: &quot;true&quot;,
    &quot;last_opened_file_path&quot;: &quot;/Users/geange/workspace/lucene-go&quot;,
    &quot;node.js.detected.package.eslint&quot;: &quot;true&quot;,
    &quot;node.js.selected.package.eslint&quot;: &quot;(autodetect)&quot;,
    &quot;nodejs_package_manager_path&quot;: &quot;npm&quot;,
    &quot;settings.editor.selected.configurable&quot;: &quot;preferences.lookFeel&quot;
  }
}</component>
  <component name="RecentsManager">
    <key name="MoveFile.RECENT_KEYS">
      <recent name="$PROJECT_DIR$/core/codecs/lucene80" />
    </key>
  </component>
  <component name="RunManager">
    <configuration name="github.com/geange/lucene-go/core/util/automaton 中的 TestFrozenIntSet_Equals/TC04_-_values_differ" type="GoTestRunConfiguration" factoryName="Go Test" temporary="true" nameIsGenerated="true">
      <module name="lucene-go" />
      <working_directory value="$PROJECT_DIR$/core/util/automaton" />
      <root_directory value="$PROJECT_DIR$" />
      <kind value="PACKAGE" />
      <package value="github.com/geange/lucene-go/core/util/automaton" />
      <directory value="$PROJECT_DIR$" />
      <filePath value="$PROJECT_DIR$" />
      <framework value="gotest" />
      <pattern value="^\QTestFrozenIntSet_Equals\E$/^\QTC04_-_values_differ\E$" />
      <method v="2" />
    </configuration>
    <recent_temporary>
      <list>
        <item itemvalue="Go 测试.github.com/geange/lucene-go/core/util/automaton 中的 TestFrozenIntSet_Equals/TC04_-_values_differ" />
      </list>
    </recent_temporary>
  </component>
  <component name="SharedIndexes">
    <attachedChunks>
      <set>
        <option value="bundled-gosdk-f466f9b0953e-3d2cccfc42a2-org.jetbrains.plugins.go.sharedIndexes.bundled-GO-252.25557.175" />
        <option value="bundled-js-predefined-d6986cc7102b-b598e85cdad2-JavaScript-GO-252.25557.175" />
      </set>
    </attachedChunks>
  </component>
  <component name="SpellCheckerSettings" RuntimeDictionaries="0" Folders="0" CustomDictionaries="0" DefaultDictionary="项目级" UseSingleDictionary="true" transferred="true" />
  <component name="TaskManager">
    <task active="true" id="Default" summary="默认任务">
      <changelist id="6d1f9e8e-b1fa-4efd-93bd-11b313d7dc2f" name="更改" comment="" />
      <created>1754326103972</created>
      <option name="number" value="Default" />
      <option name="presentableId" value="Default" />
      <updated>1754326103972</updated>
    </task>
    <servers />
  </component>
  <component name="TypeScriptGeneratedFilesManager">
    <option name="version" value="3" />
  </component>
  <component name="Vcs.Log.Tabs.Properties">
    <option name="TAB_STATES">
      <map>
        <entry key="MAIN">
          <value>
            <State />
          </value>
        </entry>
      </map>
    </option>
  </component>
  <component name="VcsManagerConfiguration">
    <MESSAGE value="fix: test fail" />
    <MESSAGE value="chore: 20250806" />
    <MESSAGE value="chore: 20250807" />
    <MESSAGE value="chore: SegmentTermsEnum" />
    <MESSAGE value="chore: TermsWriter" />
    <MESSAGE value="chore: lucene80完善format" />
    <MESSAGE value="chore: lucene80完善NormsProducer" />
    <MESSAGE value="chore: NormsProducer init" />
    <MESSAGE value="chore: IndexedDISI structure" />
    <MESSAGE value="chore: IndexedDISI" />
    <MESSAGE value="chore: NormsProducer" />
    <MESSAGE value="chore: NormsConsumer.AddNormsField" />
    <MESSAGE value="chore: NormsConsumer" />
    <MESSAGE value="chore: NormsConsumer imports" />
    <MESSAGE value="chore: update deps" />
    <MESSAGE value="chore: DocValuesConsumer.CompressedBinaryBlockWriter" />
    <MESSAGE value="chore: compress.LZ4" />
    <MESSAGE value="chore: DocValuesConsumer.writeValuesMultipleBlocks" />
    <option name="LAST_COMMIT_MESSAGE" value="chore: DocValuesConsumer.writeValuesMultipleBlocks" />
  </component>
  <component name="VgoProject">
    <settings-migrated>true</settings-migrated>
  </component>
</project>`
