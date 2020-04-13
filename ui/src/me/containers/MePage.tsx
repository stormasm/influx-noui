// Libraries
import React, {PureComponent} from 'react'
import {connect} from 'react-redux'

// Components
import {
  Panel,
  FlexBox,
  FlexDirection,
  ComponentSize,
  AlignItems,
  Grid,
  Columns,
  Page,
} from '@influxdata/clockface'
import Resources from 'src/me/components/Resources'
import Docs from 'src/me/components/Docs'
import GettingStarted from 'src/me/components/GettingStarted'
import CloudUpgradeButton from 'src/shared/components/CloudUpgradeButton'

// Utils
import {pageTitleSuffixer} from 'src/shared/utils/pageTitles'

// Types
import {AppState} from 'src/types'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface StateProps {
  me: AppState['me']
}

@ErrorHandling
export class MePage extends PureComponent<StateProps> {
  public render() {
    const {me} = this.props

    return (
      <Page titleTag={pageTitleSuffixer(['Home'])}>
        <Page.Header fullWidth={false}>
          <Page.Title title="Getting Started" />
          <CloudUpgradeButton />
        </Page.Header>
        <Page.Contents fullWidth={false} scrollable={true}>
          <Grid>
            <Grid.Row>
              <Grid.Column widthSM={Columns.Eight} widthMD={Columns.Nine}>
                <FlexBox
                  direction={FlexDirection.Column}
                  margin={ComponentSize.Small}
                  alignItems={AlignItems.Stretch}
                  stretchToFitWidth={true}
                  testID="getting-started"
                >
                  <Panel>
                    <Panel.Body>
                      <GettingStarted />
                    </Panel.Body>
                  </Panel>
                  <Docs />
                </FlexBox>
              </Grid.Column>
              <Grid.Column widthSM={Columns.Four} widthMD={Columns.Three}>
                <Resources me={me} />
              </Grid.Column>
            </Grid.Row>
          </Grid>
        </Page.Contents>
      </Page>
    )
  }
}

const mstp = (state: AppState): StateProps => {
  const {me} = state

  return {me}
}

export default connect<StateProps>(
  mstp,
  null
)(MePage)
