// Libraries
import React, {PureComponent} from 'react'
import {withRouter, WithRouterProps} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// APIs
import {client} from 'src/utils/api'

// Actions
import {dismissAllNotifications} from 'src/shared/actions/notifications'

// Components
import {ErrorHandling} from 'src/shared/decorators/errors'
import SigninForm from 'src/onboarding/components/SigninForm'
import {
  SpinnerContainer,
  TechnoSpinner,
  Panel,
  AlignItems,
  FunnelPage,
  InfluxDBCloudLogo,
} from '@influxdata/clockface'
import {RemoteDataState} from 'src/types'
import VersionInfo from 'src/shared/components/VersionInfo'

// Constants
import {CLOUD, CLOUD_SIGNIN_PATHNAME} from 'src/shared/constants'

interface State {
  status: RemoteDataState
}

interface DispatchProps {
  dismissAllNotifications: typeof dismissAllNotifications
}

type Props = WithRouterProps & DispatchProps
@ErrorHandling
class SigninPage extends PureComponent<Props, State> {
  constructor(props) {
    super(props)

    this.state = {
      status: RemoteDataState.Loading,
    }
  }
  public async componentDidMount() {
    const {allowed} = await client.setup.status()

    if (allowed) {
      this.props.router.push('/onboarding/0')
    } else if (CLOUD) {
      window.location.pathname = CLOUD_SIGNIN_PATHNAME
      return
    }

    this.setState({status: RemoteDataState.Done})
  }

  componentWillUnmount() {
    this.props.dismissAllNotifications()
  }

  public render() {
    return (
      <SpinnerContainer
        loading={this.state.status}
        spinnerComponent={<TechnoSpinner />}
      >
        <FunnelPage className="signin-page" enableGraphic={true}>
          <Panel className="signin-page--panel">
            <Panel.Body alignItems={AlignItems.Center}>
              <div className="signin-page--cubo" />
              <InfluxDBCloudLogo cloud={false} className="signin-page--logo" />
              <SigninForm />
            </Panel.Body>
            <Panel.Footer>
              <VersionInfo />
            </Panel.Footer>
          </Panel>
        </FunnelPage>
      </SpinnerContainer>
    )
  }
}

const mdtp: DispatchProps = {
  dismissAllNotifications,
}
export default connect(
  null,
  mdtp
)(withRouter(SigninPage))
